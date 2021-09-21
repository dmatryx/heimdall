package templates

import (
	"fmt"
	"testing"

	log "github.com/uswitch/heimdall/pkg/log"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	testService1 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testService1",
			Namespace:   "testNamespace",
			Annotations: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
			Selector: map[string]string{
				"app": "testApp",
			},
		},
	}

	testService2 = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testService2",
			Namespace:   "testNamespace",
			Annotations: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
			Selector: map[string]string{
				"app": "testApp",
			},
		},
	}

	testDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testApp",
			Namespace: "testNamespace",
			Labels:    map[string]string{},
			Annotations: map[string]string{
				ownerAnnotation: "testOwner",
			},
			OwnerReferences: []metav1.OwnerReference{},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: new(int32),
			Selector: &metav1.LabelSelector{},
		},
	}

	testIngressDefaultBackend = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testDefaultBackend",
			Namespace:   "testNamespace",
			Annotations: map[string]string{"com.uswitch.heimdall/5xx-rate": "0.001"},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Backend: &extensionsv1beta1.IngressBackend{
				ServiceName: "testService1",
				ServicePort: intstr.FromInt(80),
			},
		},
	}

	testIngressRuleBackend = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "testRuleBackend",
			Namespace:   "testNamespace",
			Annotations: map[string]string{"com.uswitch.heimdall/5xx-rate": "0.001"},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				extensionsv1beta1.IngressRule{
					Host: "test",
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "testService2",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

func TestIngressAnnotationsDefaultBackend(t *testing.T) {
	log.Setup(log.DEBUG_LEVEL)

	client := fake.NewSimpleClientset(testService1, testDeployment)

	template, err := NewPrometheusRuleTemplateManager("../../kube/config/templates", client)

	expr := `(
  sum(
    rate(
      nginx_ingress_controller_requests{exported_namespace="testNamespace",ingress="testDefaultBackend",status=~"5.."}[30s]
    )
  )
  /
  sum(
    rate(
      nginx_ingress_controller_requests{exported_namespace="testNamespace",ingress="testDefaultBackend"}[30s]
    )
  )
) > 0.001
`
	promrules, err := template.CreateFromIngress(testIngressDefaultBackend)
	assert.Assert(t, is.Nil(err))
	assert.Assert(t, is.Len(promrules, 1))
	assert.Equal(t, promrules[0].Spec.Groups[0].Rules[0].Expr.StrVal, expr)
}

func TestIngressAnnotationsRuleBackend(t *testing.T) {
	log.Setup(log.DEBUG_LEVEL)

	client := fake.NewSimpleClientset(testService2, testDeployment)

	template, err := NewPrometheusRuleTemplateManager("../../kube/config/templates", client)

	expr := `(
  sum(
    rate(
      nginx_ingress_controller_requests{exported_namespace="testNamespace",ingress="testRuleBackend",status=~"5.."}[30s]
    )
  )
  /
  sum(
    rate(
      nginx_ingress_controller_requests{exported_namespace="testNamespace",ingress="testRuleBackend"}[30s]
    )
  )
) > 0.001
`
	promrules, err := template.CreateFromIngress(testIngressRuleBackend)
	fmt.Println(promrules[0])

	assert.Assert(t, is.Nil(err))
	assert.Assert(t, is.Len(promrules, 1))
	assert.Equal(t, promrules[0].Spec.Groups[0].Rules[0].Expr.StrVal, expr)
}

func TestNamesMatch(t *testing.T) {
	services := []string{"test", "test", "test"}
	service := checkNamesMatch(services)
	assert.Equal(t, services[0], service)

	services = []string{"test1", "test2", "test3"}
	service = checkNamesMatch(services)
	assert.Equal(t, "", service)
}

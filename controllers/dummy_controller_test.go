package controllers

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	dummyv1alpha1 "github.com/antonmisa/dummy-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("Dummy controller", func() {
	Context("Dummy controller test", func() {

		const timeout = time.Second * 15
		const interval = time.Second * 1

		const DummyName = "test-dummy"
		const DummyNamespace = "default"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      DummyName,
				Namespace: DummyNamespace,
			},
		}

		typeNamespaceName := types.NamespacedName{Name: DummyName, Namespace: DummyNamespace}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))

			By("Setting the Image ENV VAR which stores the Operand image")
			err = os.Setenv("DUMMY_IMAGE", "interview.com/image:test")
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// TODO(user): Attention if you improve this code by adding other context test you MUST
			// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)

			By("Removing the Image ENV VAR which stores the Operand image")
			_ = os.Unsetenv("DUMMY_IMAGE")
		})

		It("should successfully reconcile a custom resource for Dummy", func() {
			By("Creating the custom resource for the Kind Dummy")
			dummy := &dummyv1alpha1.Dummy{}
			err := k8sClient.Get(ctx, typeNamespaceName, dummy)
			if err != nil && errors.IsNotFound(err) {
				// Let's mock our custom resource at the same way that we would
				// apply on the cluster the manifest under config/samples
				dummy := &dummyv1alpha1.Dummy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      DummyName,
						Namespace: DummyNamespace,
					},
					Spec: dummyv1alpha1.DummySpec{
						Message: "test message",
					},
				}

				err = k8sClient.Create(ctx, dummy)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &dummyv1alpha1.Dummy{}
				return k8sClient.Get(ctx, typeNamespaceName, found)
			}, timeout, interval).Should(Succeed())

			By("Reconciling the custom resource created")
			dummyReconciler := &DummyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = dummyReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Check Deployment created")
			Eventually(func() error {
				newCreatedDeployment := &appsv1.Deployment{}
				return k8sClient.Get(ctx, typeNamespaceName, newCreatedDeployment)
			}, timeout, interval).Should(Succeed())

			By("Checking the dummy's status message")
			Eventually(func() bool {
				found := &dummyv1alpha1.Dummy{}
				if err := k8sClient.Get(ctx, typeNamespaceName, found); err != nil {
					return false
				}
				return found.Status.SpecEcho == found.Spec.Message
			}, timeout, interval).Should(BeTrue())
		})
	})
})

func Test_unique(t *testing.T) {
	type args struct {
		s []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "Empty slice",
			args: args{[]int{}},
			want: []int{},
		},
		{
			name: "Single element slice",
			args: args{[]int{1}},
			want: []int{1},
		},
		{
			name: "Slice 6-3",
			args: args{[]int{1, 1, 1, 1, 2, 3}},
			want: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unique(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToString(t *testing.T) {
	type args struct {
		phases []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty slice",
			args: args{[]string{}},
			want: "",
		},
		{
			name: "Single element slice",
			args: args{[]string{"1"}},
			want: "1",
		},
		{
			name: "Slice 7-3",
			args: args{[]string{"3", "2", "1", "1", "1", "1", "2"}},
			want: "1, 2, 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToString(tt.args.phases); got != tt.want {
				t.Errorf("convertToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

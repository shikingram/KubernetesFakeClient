package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"testing"
)

/**
 * @Author kingram
 * @Description use fakeClient to operate deployment resources
 * @Date 2022/8/10
 **/
func TestFakeDynamicClient(t *testing.T) {
	s := runtime.NewScheme()
	err := scheme.AddToScheme(s)
	assert.Nil(t, err)
	client := fake.NewSimpleDynamicClient(s)

	var name = "my-dp"
	var ns = "ns-1"
	var rs int32= 2
	dp := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec:       appsv1.DeploymentSpec{
			Replicas: &rs,
		},
	}

	gvr := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	var dpUnstructred = &unstructured.Unstructured{}
	err = s.Convert(dp, dpUnstructred, context.Background())
	assert.Nil(t, err)

	// create deployment
	dpCreate, err := client.Resource(gvr).Namespace(ns).Create(context.Background(),dpUnstructred,metav1.CreateOptions{})
	assert.Nil(t, err)

	var newRs int32 = 1
	dp.Spec.Replicas = &newRs
	var newDpUnstructred = &unstructured.Unstructured{}
	err = s.Convert(dp, newDpUnstructred, context.Background())
	assert.Nil(t, err)

	// update deployment
	dpUpdate, err := client.Resource(gvr).Namespace(ns).Update(context.Background(),newDpUnstructred,metav1.UpdateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpUpdate)

	// get deployment
	dpGet, err := client.Resource(gvr).Namespace(ns).Get(context.Background(),name,metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpGet)

	assert.Equal(t, dpCreate.GetName(),dpGet.GetName())

	var dpGetR = &appsv1.Deployment{}
	err = s.Convert(dpGet, dpGetR, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, *dpGetR.Spec.Replicas,newRs)

	// delete deployment
	err =  client.Resource(gvr).Namespace(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	assert.Nil(t, err)
}


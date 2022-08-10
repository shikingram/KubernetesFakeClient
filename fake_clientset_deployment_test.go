package test

import (
	"context"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"log"
	"testing"
)

/**
 * @Author kingram
 * @Description use fakeClient to operate deployment resources
 * @Date 2022/8/10
 **/
func TestFakeClientSet(t *testing.T) {
	client := fake.NewSimpleClientset()
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
	// create deployment
	dpCreate,err := client.AppsV1().Deployments(ns).Create(context.Background(),dp,metav1.CreateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpCreate)

	// update deployment
	var newRs int32 = 1
	dp.Spec.Replicas = &newRs
	dpUpdate,err := client.AppsV1().Deployments(ns).Update(context.Background(),dp,metav1.UpdateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpUpdate)

	// get deployment
	dpGet,err := client.AppsV1().Deployments(ns).Get(context.Background(),name,metav1.GetOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpGet)

	assert.Equal(t, dpCreate.Name,dpGet.Name)
	assert.Equal(t, *dpGet.Spec.Replicas,newRs)

	// delete deployment
	err = client.AppsV1().Deployments(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	assert.Nil(t, err)
}

/**
 * @Author kingram
 * @Description watch event of deployment
 * @Date 2022/8/10
 **/
func TestClientSetWatch(t *testing.T) {
	client := fake.NewSimpleClientset()

	stopChain := make(chan struct{})

	var name = "my-dp"
	var ns = "ns-1"
	var rs int32= 2
	dp := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"dp":"1",
			},
		},
		Spec:       appsv1.DeploymentSpec{
			Replicas: &rs,
		},
	}

	dpWatch,err := client.AppsV1().Deployments(ns).Watch(context.Background(),metav1.ListOptions{
		LabelSelector:        "dp=1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, dpWatch)

	go func() {
		for {
			select {
			case e,ok:= <- dpWatch.ResultChan():
				if ok {
					assert.Equal(t,e.Type,watch.Added)
					log.Printf("Got Event:%+v \n",e)
					stopChain <- struct{}{}
				}
			}
		}
	}()

	// create deployment
	dpCreate,err := client.AppsV1().Deployments(ns).Create(context.Background(),dp,metav1.CreateOptions{})
	assert.Nil(t, err)
	assert.NotNil(t, dpCreate)

	<- stopChain
}

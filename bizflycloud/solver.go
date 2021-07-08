package bizflycloud

import (
	"context"
	"fmt"

	"github.com/bizflycloud/gobizfly"
	"github.com/jetstack/cert-manager/pkg/acme/webhook"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func NewSolver() webhook.Solver {
	return &bizflycloudDNSProviderSolver{}
}

type bizflycloudDNSProviderSolver struct {
	client *kubernetes.Clientset
}

func (s *bizflycloudDNSProviderSolver) Name() string {
	return "bizflycloud"
}

type bizflycloudDNSProviderConfig struct {
}

func (s *bizflycloudDNSProviderSolver) newClientFromChallenge(ch *v1alpha1.ChallengeRequest) (*Client, error) {

	client, err := newClient()
	if err != nil {
		return nil, fmt.Errorf("new dns client error: %v", err)
	}

	return client, nil
}

func (s *bizflycloudDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {

	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	s.client = cl

	klog.Infof("Initialize Successed")

	return nil
}

func (s *bizflycloudDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Presenting txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := s.newClientFromChallenge(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	zoneID, err := client.domainNameToZoneID(ch.ResolvedZone)
	if err != nil {
		return err
	}

	values := []string{ch.Key}

	recordName := ch.ResolvedFQDN[0 : len(ch.ResolvedFQDN)-len(ch.ResolvedZone)]
	if last := len(recordName) - 1; last >= 0 && recordName[last] == '.' {
		recordName = recordName[:last]
	}

	createRequest := &gobizfly.CreateNormalRecordPayload{
		BaseCreateRecordPayload: gobizfly.BaseCreateRecordPayload{
			Name: recordName,
			Type: "TXT",
			TTL:  3600,
		},
		Data: values,
	}

	_, err = client.dnsc.DNS.CreateRecord(
		context.Background(),
		zoneID,
		createRequest,
	)
	if err != nil {
		return err
	}

	klog.Infof("Presented txt record %v", ch.ResolvedFQDN)
	return nil
}

func (s *bizflycloudDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Cleaning up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := s.newClientFromChallenge(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	records, ID, err := client.findTxtRecord(ch.ResolvedZone, ch.ResolvedFQDN)
	if err != nil {
		return err
	}

	if ID == "" {
		klog.Infof("No TXT record exist %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
		return nil
	}

	for range records {
		err = client.dnsc.DNS.DeleteRecord(context.Background(), ID)
		if err != nil {
			return err
		}
	}

	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}

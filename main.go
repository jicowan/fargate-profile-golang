package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/satori/go.uuid"
	"time"
)

type profile struct {
	profileName *string
	clusterName *string
	arn *string
}

func main() {
	fp := flag.String("FargateProfileName", "GoProfile", "")
	cluster := flag.String("ClusterName", "riverrun", "")
	arn := flag.String("RoleArn", "arn:aws:iam::820537372947:role/AmazonEKSFargatePodExecutionRole", "")

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(endpoints.UsEast2RegionID)}))
	p := profile{
		profileName: fp,
		clusterName:  cluster,
		arn:    arn,
	}
	p.create(sess)
	if p.isReady(sess) {
		p.deletes(sess)
	}
}

func (p *profile) isReady(s *session.Session) bool {
	svc := eks.New(s)
	for true {
		output, err := svc.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
			ClusterName:        p.clusterName,
			FargateProfileName: p.profileName,
		})
		if *output.FargateProfile.Status == "ACTIVE" {
			return true
		} else if err != nil {
			fmt.Printf("Something went horribly wrong %v\n", err)
			return false
		} else {
			fmt.Printf("The current status is: %v. Sleeping for 1 second.\n", *output.FargateProfile.Status)
			time.Sleep(1 * time.Second)
		}
	}
	return false
}

func (p *profile) create(s *session.Session) {
	svc := eks.New(s)
	subnets := []*string{
		aws.String("subnet-0b2537287adffd2b7"),
		aws.String("subnet-06209a8281b744e3e"),
		aws.String("subnet-0540ebac0f3bad2fe"),
	}
	selector := eks.FargateProfileSelector{
		Namespace: aws.String("fargate"),
		Labels: map[string]*string{
			"foo": aws.String("bar"),
		},
	}
	selectors := []*eks.FargateProfileSelector{&selector}
	_, err := svc.CreateFargateProfile(&eks.CreateFargateProfileInput{
			ClientRequestToken: aws.String(uuid.Must(uuid.NewV4()).String()),
			ClusterName: p.clusterName,
			PodExecutionRoleArn: p.arn,
			FargateProfileName: p.profileName,
			Selectors: selectors,
			Subnets: subnets,
		},
	)
	if err != nil {
		panic("Something went horribly wrong!")
	}
}

func (p *profile) deletes(s *session.Session) {
	svc := eks.New(s)
	_, err := svc.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
		ClusterName: p.clusterName,
		FargateProfileName: p.profileName,
		},
	)
	if err != nil {
		panic("Something went horribly wrong!")
	}
}

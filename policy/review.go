package policy

import (
	"log"
	"regexp"

	"github.com/tczekajlo/kir/etcd"
	"github.com/tczekajlo/kir/pb"
	"github.com/tczekajlo/kir/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Review makes image review and returns data of matched rule
func Review(req *types.ImageReview) *types.ImageReviewResponse {
	meta := metav1.TypeMeta{
		Kind:       req.TypeMeta.Kind,
		APIVersion: req.TypeMeta.APIVersion,
	}

	etcd := etcd.Client{}
	etcd.New()

	rules, err := etcd.GetAll(0)
	if err != nil {
		log.Panic(err)
	}
	defer etcd.Client.Close()

	for _, rule := range rules.Rule {
		if checkRule(rule, req) {
			return &types.ImageReviewResponse{
				TypeMeta: meta,
				Status: types.ImageReviewStatus{
					Allowed: rule.Allowed,
					Reason:  rule.Reason,
				},
			}
		}
	}

	// prepare response
	return &types.ImageReviewResponse{
		TypeMeta: meta,
		Status: types.ImageReviewStatus{
			Allowed: false,
			Reason:  "Cannot match to any rule",
		},
	}

}

// checkRules checks if rule fulfill conditions.
func checkRule(rule *pb.Rule, req *types.ImageReview) bool {
	image, annotations, namespace := false, false, false

	for _, reqContainer := range req.Spec.Containers {
		for _, container := range rule.Containers {
			if matched, _ := regexp.MatchString(container.Image, reqContainer.Image); matched {
				image = true
				break
			}
		}
	}

	for reqKey, reqValue := range req.Spec.Annotations {
		for key, value := range rule.Annotations {
			if matched, _ := regexp.MatchString(key+"="+value, reqKey+"="+reqValue); matched {
				annotations = true
				break
			}

		}
	}
	// in the case when in request and in rule is lack of annotations then return true
	if len(req.Spec.Annotations) == 0 && len(rule.Annotations) == 0 {
		annotations = true
	}

	if matched, _ := regexp.MatchString(rule.Namespace, req.Spec.Namespace); matched {
		namespace = true
	}

	return (image && annotations && namespace)
}

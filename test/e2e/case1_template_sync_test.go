// Copyright (c) 2020 Red Hat, Inc.

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/governance-policy-propagator/pkg/controller/common"
	"github.com/open-cluster-management/governance-policy-propagator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const case1PolicyName string = "default.case1-test-policy"
const case1PolicyYaml string = "../resources/case1_template_sync/case1-test-policy.yaml"
const cast1TrustedContainerPolicyName string = "case1-test-policy-trustedcontainerpolicy"

var _ = Describe("Test spec sync", func() {
	BeforeEach(func() {
		By("Creating a policy on managed cluster in ns:" + testNamespace)
		utils.Kubectl("apply", "-f", case1PolicyYaml, "-n", testNamespace)
		plc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(plc).NotTo(BeNil())
	})
	AfterEach(func() {
		By("Deleting a policy on managed cluster in ns:" + testNamespace)
		utils.Kubectl("delete", "-f", case1PolicyYaml, "-n", testNamespace)
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
	})
	It("should create policy template on managed cluster", func() {
		By("Checking the trustedcontainerpolicy CR")
		yamlTrustedPlc := utils.ParseYaml("../resources/case1_template_sync/case1-trusted-container-policy.yaml")
		Eventually(func() interface{} {
			trustedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrTrustedContainerPolicy, cast1TrustedContainerPolicyName, testNamespace, true, defaultTimeoutSeconds)
			return trustedPlc.Object["spec"]
		}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlTrustedPlc.Object["spec"]))
	})
	It("should override remediationAction in spec", func() {
		By("Patching policy remediationAction=enforce")
		plc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		plc.Object["spec"].(map[string]interface{})["remediationAction"] = "enforce"
		plc, err := clientManagedDynamic.Resource(gvrPolicy).Namespace("managed").Update(plc, metav1.UpdateOptions{})
		Expect(err).To(BeNil())
		Expect(plc.Object["spec"].(map[string]interface{})["remediationAction"]).To(Equal("enforce"))
		By("Checking template policy remediationAction")
		yamlTrustedPlc := utils.ParseYaml("../resources/case1_template_sync/case1-trusted-container-policy-enforce.yaml")
		Eventually(func() interface{} {
			trustedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrTrustedContainerPolicy, cast1TrustedContainerPolicyName, testNamespace, true, defaultTimeoutSeconds)
			return trustedPlc.Object["spec"]
		}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlTrustedPlc.Object["spec"]))
	})
	It("should still override remediationAction in spec when there is no remediationAction", func() {
		By("Updating policy with no remediationAction")
		utils.Kubectl("apply", "-f", "../resources/case1_template_sync/case1-test-policy-no-remediation.yaml", "-n", testNamespace)
		By("Checking template policy remediationAction")
		yamlTrustedPlc := utils.ParseYaml("../resources/case1_template_sync/case1-trusted-container-policy-enforce.yaml")
		Eventually(func() interface{} {
			trustedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrTrustedContainerPolicy, cast1TrustedContainerPolicyName, testNamespace, true, defaultTimeoutSeconds)
			return trustedPlc.Object["spec"]
		}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlTrustedPlc.Object["spec"]))
	})
	It("should contains labels from parent policy", func() {
		By("Checking labels of template policy")
		plc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		trustedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrTrustedContainerPolicy, cast1TrustedContainerPolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(plc.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})[common.ClusterNameLabel]).To(
			utils.SemanticEqual(trustedPlc.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})[common.ClusterNameLabel]))
		Expect(plc.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})[common.ClusterNamespaceLabel]).To(
			utils.SemanticEqual(trustedPlc.Object["metadata"].(map[string]interface{})["labels"].(map[string]interface{})[common.ClusterNamespaceLabel]))
	})
	It("should delete template policy on managed cluster", func() {
		By("Deleting parent policy")
		utils.Kubectl("delete", "-f", case1PolicyYaml, "-n", testNamespace)
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		By("Checking the existence of template policy")
		utils.GetWithTimeout(clientManagedDynamic, gvrTrustedContainerPolicy, cast1TrustedContainerPolicyName, testNamespace, false, defaultTimeoutSeconds)
	})
})

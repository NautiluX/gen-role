package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/NautiluX/gen-role/pkg/plugin"
)

var _ = Describe("Plugin", func() {
	Context("ParsePerms", func() {
		var (
			inputMatches []string
		)
		Context("When reading resources", func() {
			It("returns correct permissions for oc get pods -n openshift-console", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/api/v1/namespaces/openshift-console/pods?limit=500"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("list"))
				Expect(perms.Namespace).To(Equal("openshift-console"))
				Expect(perms.Resource).To(Equal("pods"))
				Expect(perms.Api).To(Equal("v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

			It("returns correct permissions for oc get pods -n openshift-console", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/api/v1/namespaces/openshift-console/pods/console-6ccf9997ff-qsnkf"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("get"))
				Expect(perms.Namespace).To(Equal("openshift-console"))
				Expect(perms.Resource).To(Equal("pods"))
				Expect(perms.Api).To(Equal("v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

			It("returns correct permissions for oc get deployment -n openshift-console", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/apis/apps/v1/namespaces/openshift-console/deployments?limit=500"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("list"))
				Expect(perms.Namespace).To(Equal("openshift-console"))
				Expect(perms.Resource).To(Equal("deployments"))
				Expect(perms.Api).To(Equal("apps/v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

			It("returns correct permissions for oc get deployment console -n openshift-console", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/apis/apps/v1/namespaces/openshift-console/deployments/console"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("get"))
				Expect(perms.Namespace).To(Equal("openshift-console"))
				Expect(perms.Resource).To(Equal("deployments"))
				Expect(perms.Api).To(Equal("apps/v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

			It("returns correct permissions for oc get deployment -A", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/apis/apps/v1/deployments?limit=500"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("list"))
				Expect(perms.Namespace).To(Equal(""))
				Expect(perms.Resource).To(Equal("deployments"))
				Expect(perms.Api).To(Equal("apps/v1"))
				Expect(perms.Namespaced).To(BeFalse())
			})

			It("returns correct permissions for oc get pod -A", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/api/v1/pods?limit=500"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("list"))
				Expect(perms.Namespace).To(Equal(""))
				Expect(perms.Resource).To(Equal("pods"))
				Expect(perms.Api).To(Equal("v1"))
				Expect(perms.Namespaced).To(BeFalse())
			})

			It("returns correct permissions for oc exec -ti curl curl google.com", func() {
				inputMatches = []string{"", "POST", "https://api.crc.testing:6443/api/v1/namespaces/default/pods/curl/exec?command=curl&command=google.com&container=curl&stderr=true&stdin=true&stdout=true"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("post"))
				Expect(perms.Namespace).To(Equal("default"))
				Expect(perms.Resource).To(Equal("pods"))
				Expect(perms.Api).To(Equal("v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

			It("returns correct permissions for watching resources", func() {
				inputMatches = []string{"", "GET", "https://api.crc.testing:6443/api/v1/namespaces/default/pods?fieldSelector=metadata.name%3Dcurl&resourceVersion=107674&watch=true"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("watch"))
				Expect(perms.Namespace).To(Equal("default"))
				Expect(perms.Resource).To(Equal("pods"))
				Expect(perms.Api).To(Equal("v1"))
				Expect(perms.Namespaced).To(BeTrue())
			})

		})
		Context("When modifying resources", func() {
			It("returns correct permissions for clusterrolebinding deletion", func() {
				inputMatches = []string{"", "DELETE", "https://api.crc.testing:6443/apis/rbac.authorization.k8s.io/v1/clusterrolebindings/system:openshift:scc:kubeadmin"}
				perms, _ := ParsePerms(inputMatches)
				Expect(perms.Verb).To(Equal("delete"))
				Expect(perms.Namespace).To(Equal(""))
				Expect(perms.Resource).To(Equal("clusterrolebindings"))
				Expect(perms.Api).To(Equal("rbac.authorization.k8s.io/v1"))
				Expect(perms.Namespaced).To(BeFalse())
			})
		})

	})

})

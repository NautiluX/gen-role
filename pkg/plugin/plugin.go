package plugin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	v1 "k8s.io/api/rbac/v1"

	"github.com/ghodss/yaml"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type Perm struct {
	Verb       string
	Resource   string
	Api        string
	Namespace  string
	Namespaced bool
}

func RunPlugin(configFlags *genericclioptions.ConfigFlags) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	_, err = kubernetes.NewForConfig(config) // clientset
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	kubectlBin := os.Getenv("_")
	kubectlArgs := os.Args[1:]
	kubectlArgs = append(kubectlArgs, "-v6")

	fmt.Printf("Running %s \"%s\"\n", kubectlBin, strings.Join(kubectlArgs, "\" \""))
	cmd := exec.Command(kubectlBin, kubectlArgs...)

	var buf1 bytes.Buffer
	w := io.MultiWriter(&buf1, os.Stderr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = w
	cmd.Run()

	scanner := bufio.NewScanner(&buf1)
	fmt.Println("Matched requests:")

	perms := []Perm{}
	regexRequest := regexp.MustCompilePOSIX(".*(GET|DELETE|POST|PATCH|PUT) https?://.*:[0-9]+/(.*) 2.. OK.*")
	for scanner.Scan() {
		line := scanner.Text()
		matches := regexRequest.FindStringSubmatch(line)
		if len(matches) > 0 {
			fmt.Println(line)
			//fmt.Printf("Matches: %d ", len(matches) )
			//for _, m := range(matches {
			//fmt.Println(m)
			//}
			perm, ok := ParsePerms(matches)
			if ok {
				perms = append(perms, perm)
			}
		}
	}

	//fmt.Printf("perms: %v\n", perms)

	role := GenerateRole(perms)
	if len(role.Rules) > 0 {
		roleJsonBytes, err := yaml.Marshal(role)
		if err != nil {
			return fmt.Errorf("error generating Role yaml: %v", err)
		}
		fmt.Printf("Role:\n %s\n", string(roleJsonBytes))
		ioutil.WriteFile("gen-role.yaml", roleJsonBytes, 0644)
	}
	clusterRole := GenerateClusterRole(perms)
	if len(clusterRole.Rules) > 0 {

		clusterRoleJsonBytes, err := yaml.Marshal(clusterRole)
		if err != nil {
			return fmt.Errorf("error generating ClusterRole yaml: %v", err)
		}
		fmt.Printf("ClusterRole:\n %s\n", string(clusterRoleJsonBytes))
		ioutil.WriteFile("gen-cluster-role.yaml", clusterRoleJsonBytes, 0644)
	}
	return nil
}

func ParsePerms(matches []string) (p Perm, success bool) {
	p.Verb = strings.ToLower(matches[1])

	regexNamespace := regexp.MustCompilePOSIX("apis?/(.*)/namespaces/([^/]+)/([^/?]+)(/[^/?]*)?(\\?.*)?")
	urlMatches := regexNamespace.FindStringSubmatch(matches[2])
	//fmt.Printf("Matches: %d\n", len(urlMatches))
	//for _, m := range urlMatches {
	//	fmt.Println(m)
	//}
	if len(urlMatches) == 6 {
		p.Api = urlMatches[1]
		p.Namespace = urlMatches[2]
		p.Resource = urlMatches[3]
		resourceName := urlMatches[4]
		p.Namespaced = true
		// parameters = urlMatches[5]
		if resourceName == "" && p.Verb == "get" {
			p.Verb = "list"
		}
		return p, true
	}

	regexNotNamespaced := regexp.MustCompilePOSIX("apis?/([^/?]+(/[^/?]+)?)/([^/?]+)(/[^/?]+)?")
	urlMatches = regexNotNamespaced.FindStringSubmatch(matches[2])
	//fmt.Printf("Matches: %d\n", len(urlMatches))
	//for _, m := range urlMatches {
	//	fmt.Println(m)
	//}
	if len(urlMatches) == 5 {
		p.Api = urlMatches[1]
		p.Resource = urlMatches[3]
		p.Namespaced = false
		resourceName := urlMatches[4]
		if resourceName == "" && p.Verb == "get" {
			p.Verb = "list"
		}
		return p, true
	}

	return p, false
}

func GenerateRole(perms []Perm) (role v1.Role) {
	yamlFile := getFileContentIfExists("gen-role.yaml")
	if yamlFile != nil {
		err := yaml.Unmarshal(yamlFile, &role)
		if err != nil {
			fmt.Printf("Warning: Can't parse yaml from gen-role.yaml, starting with a fresh Role: %v\n", err)
			role = v1.Role{}
		}
		fmt.Println("Adding permissions to role definition in gen-role.yaml")
	}

	role.Name = "gen-role-generated-role"
	role.Namespace = "gen-role-generated-role"
	for _, p := range perms {
		if !p.Namespaced {
			continue
		}
		role.Rules = mergePerm(role.Rules, p)
	}

	return
}

func getFileContentIfExists(filename string) []byte {
	_, error := os.Stat(filename)
	if os.IsNotExist(error) {
		return nil
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("can't read file %s: %v", filename, err)
	}
	return file
}

func GenerateClusterRole(perms []Perm) (role v1.ClusterRole) {
	yamlFile := getFileContentIfExists("gen-cluster-role.yaml")
	if yamlFile != nil {
		err := yaml.Unmarshal(yamlFile, &role)
		if err != nil {
			fmt.Printf("Warning: Can't parse yaml from gen-cluster-role.yaml, starting with a fresh ClusterRole: %v\n", err)
			role = v1.ClusterRole{}
		}
		fmt.Println("Adding permissions to cluster-role definition in gen-cluster-role.yaml")
	}
	role.Name = "gen-role-generated-clusterrole"
	role.Namespace = "gen-role-generated-clusterrole"
	for _, p := range perms {
		if p.Namespaced {
			continue
		}
		role.Rules = mergePerm(role.Rules, p)
	}

	return
}

func mergePerm(rules []v1.PolicyRule, p Perm) []v1.PolicyRule {
	api := p.Api
	if p.Api == "v1" {
		api = ""
	}
	for i, rule := range rules {
		if contains(rule.APIGroups, api) && contains(rule.ResourceNames, p.Resource) {
			if !contains(rule.Verbs, p.Verb) {
				rules[i].Verbs = append(rules[i].Verbs, p.Verb)
			}
			return rules
		}
	}
	return append(rules, v1.PolicyRule{
		APIGroups:     []string{api},
		Verbs:         []string{p.Verb},
		ResourceNames: []string{p.Resource},
	})
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

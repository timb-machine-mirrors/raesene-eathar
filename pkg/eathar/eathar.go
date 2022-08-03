package eathar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//This needs to be exported to work with the JSON marshalling
// omitempty thing is there as container won't always be relevant (e.g. hostPID)
type Finding struct {
	Check        string
	Namespace    string
	Pod          string
	Container    string   `json:",omitempty"`
	Capabilities []string `json:",omitempty"`
	Hostport     int      `json:",omitempty"`
}

func Hostnet(options *pflag.FlagSet) {
	var hostnetcont []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//Debugging command
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {

		if pod.Spec.HostNetwork {
			// We set the container to blank as there's no container for this finding
			p := Finding{Check: "hostnet", Namespace: pod.Namespace, Pod: pod.Name, Container: ""}
			//fmt.Printf("Namespace %s - Pod %s is using Host networking\n", p.namespace, p.pod)
			hostnetcont = append(hostnetcont, p)
		}
	}
	report(hostnetcont, jsonrep)
}

func Hostpid(options *pflag.FlagSet) {
	var hostpidcont []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//Debugging command
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	for _, pod := range pods.Items {

		if pod.Spec.HostPID {
			p := Finding{Check: "hostpid", Namespace: pod.Namespace, Pod: pod.Name, Container: ""}
			//fmt.Printf("Namespace %s - Pod %s is using Host PID\n", p.namespace, p.pod)
			hostpidcont = append(hostpidcont, p)
		}
	}
	report(hostpidcont, jsonrep)
}

func Hostipc(options *pflag.FlagSet) {
	var hostipccont []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//Debugging command
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	for _, pod := range pods.Items {

		if pod.Spec.HostIPC {
			p := Finding{Check: "hostipc", Namespace: pod.Namespace, Pod: pod.Name, Container: ""}
			//fmt.Printf("Namespace %s - Pod %s is using Host PID\n", p.namespace, p.pod)
			hostipccont = append(hostipccont, p)
		}
	}
	report(hostipccont, jsonrep)
}

func AllowPrivEsc(options *pflag.FlagSet) {
	var allowprivesccont []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//Debugging command
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			// Logic here is if there's no security context, or there is a security context and no mention of allow privilege escalation then the default is true
			// We don't catch the case of someone explicitly setting it to true, but that seems unlikely
			allowPrivilegeEscalation := (container.SecurityContext == nil) || (container.SecurityContext != nil && container.SecurityContext.AllowPrivilegeEscalation == nil)
			if allowPrivilegeEscalation {
				p := Finding{Check: "allowprivesc", Namespace: pod.Namespace, Pod: pod.Name, Container: container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container: %s does not block privilege escalation\n", p.namespace, p.pod, p.container)
				allowprivesccont = append(allowprivesccont, p)
			}
		}
		for _, init_container := range pod.Spec.InitContainers {
			// Logic here is if there's no security context, or there is a security context and no mention of allow privilege escalation then the default is true
			// We don't catch the case of someone explicitly setting it to true, but that seems unlikely
			allowPrivilegeEscalation := (init_container.SecurityContext == nil) || (init_container.SecurityContext != nil && init_container.SecurityContext.AllowPrivilegeEscalation == nil)
			if allowPrivilegeEscalation {
				p := Finding{Check: "allowprivesc", Namespace: pod.Namespace, Pod: pod.Name, Container: init_container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container: %s does not block privilege escalation\n", p.namespace, p.pod, p.container)
				allowprivesccont = append(allowprivesccont, p)
			}
		}
		for _, eph_container := range pod.Spec.EphemeralContainers {
			// Logic here is if there's no security context, or there is a security context and no mention of allow privilege escalation then the default is true
			// We don't catch the case of someone explicitly setting it to true, but that seems unlikely
			allowPrivilegeEscalation := (eph_container.SecurityContext == nil) || (eph_container.SecurityContext != nil && eph_container.SecurityContext.AllowPrivilegeEscalation == nil)
			if allowPrivilegeEscalation {
				p := Finding{Check: "allowprivesc", Namespace: pod.Namespace, Pod: pod.Name, Container: eph_container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container: %s does not block privilege escalation\n", p.namespace, p.pod, p.container)
				allowprivesccont = append(allowprivesccont, p)
			}
		}
	}

	report(allowprivesccont, jsonrep)
}

func Privileged(options *pflag.FlagSet) {
	var privcont []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	//Debugging command
	//fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			// if you try to check privileged for nil on it's own, it doesn't work you need to check security context too
			privileged_container := container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged
			if privileged_container {
				// So we create a new privileged struct from our matching container
				p := Finding{Check: "privileged", Namespace: pod.Namespace, Pod: pod.Name, Container: container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container  : %s is running as privileged \n", p.namespace, p.pod, p.container)
				//And we append it to our slice of all our privileged containers
				privcont = append(privcont, p)
			}
		}
		for _, init_container := range pod.Spec.InitContainers {
			// if you try to check privileged for nil on it's own, it doesn't work you need to check security context too
			privileged_container := init_container.SecurityContext != nil && init_container.SecurityContext.Privileged != nil && *init_container.SecurityContext.Privileged
			if privileged_container {
				// So we create a new privileged struct from our matching container
				p := Finding{Check: "privileged", Namespace: pod.Namespace, Pod: pod.Name, Container: init_container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container  : %s is running as privileged \n", p.namespace, p.pod, p.container)
				//And we append it to our slice of all our privileged containers
				privcont = append(privcont, p)
			}
		}
		for _, eph_container := range pod.Spec.EphemeralContainers {
			// if you try to check privileged for nil on it's own, it doesn't work you need to check security context too
			privileged_container := eph_container.SecurityContext != nil && eph_container.SecurityContext.Privileged != nil && *eph_container.SecurityContext.Privileged
			if privileged_container {
				// So we create a new privileged struct from our matching container
				p := Finding{Check: "privileged", Namespace: pod.Namespace, Pod: pod.Name, Container: eph_container.Name}
				//fmt.Printf("Namespace: %s - Pod: %s - Container  : %s is running as privileged \n", p.namespace, p.pod, p.container)
				//And we append it to our slice of all our privileged containers
				privcont = append(privcont, p)
			}
		}
	}
	// Just to prove our slice is working
	report(privcont, jsonrep)
}

func AddedCapabilities(options *pflag.FlagSet) {
	var capadded []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			cap_added := container.SecurityContext != nil && container.SecurityContext.Capabilities != nil && container.SecurityContext.Capabilities.Add != nil
			if cap_added {
				//Need to convert the capabilities struct to strings, I think.
				var added_caps []string
				for _, cap := range container.SecurityContext.Capabilities.Add {
					added_caps = append(added_caps, string(cap))
				}
				p := Finding{Check: "Added Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: container.Name, Capabilities: added_caps}
				capadded = append(capadded, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}

		for _, init_container := range pod.Spec.InitContainers {
			cap_added := init_container.SecurityContext != nil && init_container.SecurityContext.Capabilities != nil && init_container.SecurityContext.Capabilities.Add != nil
			if cap_added {
				//Need to convert the capabilities struct to strings, I think.
				var added_caps []string
				for _, cap := range init_container.SecurityContext.Capabilities.Add {
					added_caps = append(added_caps, string(cap))
				}
				p := Finding{Check: "Added Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: init_container.Name, Capabilities: added_caps}
				capadded = append(capadded, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}

		for _, eph_container := range pod.Spec.EphemeralContainers {
			cap_added := eph_container.SecurityContext != nil && eph_container.SecurityContext.Capabilities != nil && eph_container.SecurityContext.Capabilities.Add != nil
			if cap_added {
				//Need to convert the capabilities struct to strings, I think.
				var added_caps []string
				for _, cap := range eph_container.SecurityContext.Capabilities.Add {
					added_caps = append(added_caps, string(cap))
				}
				p := Finding{Check: "Added Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: eph_container.Name, Capabilities: added_caps}
				capadded = append(capadded, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}
	}
	report(capadded, jsonrep)
}

func DroppedCapabilities(options *pflag.FlagSet) {
	var capdropped []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			cap_dropped := container.SecurityContext != nil && container.SecurityContext.Capabilities != nil && container.SecurityContext.Capabilities.Drop != nil
			if cap_dropped {
				//Need to convert the capabilities struct to strings, I think.
				var dropped_caps []string
				for _, cap := range container.SecurityContext.Capabilities.Drop {
					dropped_caps = append(dropped_caps, string(cap))
				}
				p := Finding{Check: "Dropped Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: container.Name, Capabilities: dropped_caps}
				capdropped = append(capdropped, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}

		for _, init_container := range pod.Spec.InitContainers {
			cap_dropped := init_container.SecurityContext != nil && init_container.SecurityContext.Capabilities != nil && init_container.SecurityContext.Capabilities.Drop != nil
			if cap_dropped {
				//Need to convert the capabilities struct to strings, I think.
				var dropped_caps []string
				for _, cap := range init_container.SecurityContext.Capabilities.Drop {
					dropped_caps = append(dropped_caps, string(cap))
				}
				p := Finding{Check: "Dropped Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: init_container.Name, Capabilities: dropped_caps}
				capdropped = append(capdropped, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}

		for _, eph_container := range pod.Spec.EphemeralContainers {
			cap_dropped := eph_container.SecurityContext != nil && eph_container.SecurityContext.Capabilities != nil && eph_container.SecurityContext.Capabilities.Drop != nil
			if cap_dropped {
				//Need to convert the capabilities struct to strings, I think.
				var dropped_caps []string
				for _, cap := range eph_container.SecurityContext.Capabilities.Drop {
					dropped_caps = append(dropped_caps, string(cap))
				}
				p := Finding{Check: "Dropped Capabilities", Namespace: pod.Namespace, Pod: pod.Name, Container: eph_container.Name, Capabilities: dropped_caps}
				capdropped = append(capdropped, p)
				//debugging command
				//fmt.Println(strings.Join(added_caps[:], ","))
			}
		}
	}
	report(capdropped, jsonrep)
}

func HostPorts(options *pflag.FlagSet) {
	var hostports []Finding
	kubeconfig, _ := options.GetString("kubeconfig")
	jsonrep, _ := options.GetBool("jsonrep")
	clientset := connectToCluster(kubeconfig)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			//Does the container have ports specified
			cports := container.Ports != nil
			if cports {
				for _, port := range container.Ports {
					// Is the port a host port
					if port.HostPort != 0 {
						p := Finding{Check: "Host Ports", Namespace: pod.Namespace, Pod: pod.Name, Container: container.Name, Hostport: int(port.HostPort)}
						hostports = append(hostports, p)
					}
				}
			}
		}
		for _, init_container := range pod.Spec.InitContainers {
			//Does the container have ports specified
			cports := init_container.Ports != nil
			if cports {
				for _, port := range init_container.Ports {
					// Is the port a host port
					if port.HostPort != 0 {
						p := Finding{Check: "Host Ports", Namespace: pod.Namespace, Pod: pod.Name, Container: init_container.Name, Hostport: int(port.HostPort)}
						hostports = append(hostports, p)
					}
				}
			}
		}
		for _, eph_container := range pod.Spec.EphemeralContainers {
			//Does the container have ports specified
			cports := eph_container.Ports != nil
			if cports {
				for _, port := range eph_container.Ports {
					// Is the port a host port
					if port.HostPort != 0 {
						p := Finding{Check: "Host Ports", Namespace: pod.Namespace, Pod: pod.Name, Container: eph_container.Name, Hostport: int(port.HostPort)}
						hostports = append(hostports, p)
					}
				}
			}
		}
	}
	report(hostports, jsonrep)
}

// This is our function for connecting to the cluster
func connectToCluster(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientset
}

func report(f []Finding, jsonrep bool) {
	if !jsonrep {
		if f != nil {
			fmt.Printf("Findings for the %s check\n", f[0].Check)
			for _, i := range f {
				switch i.Check {
				case "hostpid", "hostnet":
					fmt.Printf("namespace %s : pod %s\n", i.Namespace, i.Pod)
				case "privileged", "allowprivesc":
					fmt.Printf("namespace %s : pod %s : container %s\n", i.Namespace, i.Pod, i.Container)
				case "Added Capabilities":
					fmt.Printf("namespace %s : pod %s : container %s added %s capabilities\n", i.Namespace, i.Pod, i.Container, strings.Join(i.Capabilities[:], ","))
				case "Dropped Capabilities":
					fmt.Printf("namespace %s : pod %s : container %s dropped %s capabilities\n", i.Namespace, i.Pod, i.Container, strings.Join(i.Capabilities[:], ","))
				case "Host Ports":
					fmt.Printf("namespace %s : pod %s : container %s : port %d\n", i.Namespace, i.Pod, i.Container, i.Hostport)
				}
			}
		} else {
			fmt.Println("No findings!")
		}
	} else {

		js, err := json.MarshalIndent(f, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(js))
	}

}

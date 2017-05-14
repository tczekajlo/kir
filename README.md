# KIR

The KIR (Kubernetes Image Reviewer) is a tool to makes review of container's image used in PODs.

## Requirements

- etcdv3

## Configuration

### Kubernetes

You can find [here](https://kubernetes.io/docs/admin/admission-controllers/#imagepolicywebhook) details information how looks ImagePolicyWebhook configuration for Kubernetes. As server endpoint, you have to use `/api/v1/review` path.

### KIR

In order to a configuration, you can use configuration file or flags. You can find [here](https://github.com/tczekajlo/kir/tree/master/examples/kir_config.yaml) an example of the configuration file.

Kubernetes' ImagePolicyWebhook requires HTTPS. Below an example how to run the server with enabled TLS configuration.

```
:~# kir server --tls-enabled --tls-key-file=apiserver.key --tls-cert-file=apiserver.crt --tls-require-and-verify-client-cert --tls-cacert-file=ca.crt 
 __  ___  __  .______      
|  |/  / |  | |   _  \     
|  '  /  |  | |  |_)  |    
|    <   |  | |      /     
|  .  \  |  | |  |\  \---.
|__|\__\ |__| | _| ._____| 

2017-05-14 09:18:32.901041 I | 12517 :8080

```

## Examples of usage

### Add new rule

In this example, we add the rule which says that all images which are matched to `^httpd:2.2.*$` or `^nginx$` regex will be banned.

```
:~# kir add --image ^httpd:2.2.*$,^nginx$ --name banned --namespace ^default$ --reason "I don't like this images"

# Show rules
:~# kir get banned
    NAME     NAMESPACE          IMAGE                           ANNOTATIONS                    ALLOWED  
 ---------- ----------- ---------------------- ---------------------------------------------- --------- 
  banned     ^default$   ^httpd:2.2.*$          none=none                                      false    
                         ^nginx$                                                                        

```
You can add a rule from file as well, e.g. `kir add -f examples/rules/banned.yaml`. In the case when a rule already exists and you want to override existing one you can use `--override` flag.

Now, every container which uses `nginx` image will be banned. Below the example.

```
:~# kubectl get rs nginx-2371676037
Name:       nginx-2371676037
Namespace:  default
Image(s):   nginx
Selector:   pod-template-hash=2371676037,run=nginx
Labels:     pod-template-hash=2371676037
        run=nginx
Replicas:   0 current / 1 desired
Pods Status:    0 Running / 0 Waiting / 0 Succeeded / 0 Failed
No volumes.
Events:
  FirstSeen LastSeen    Count   From                SubObjectPath   Type        Reason      Message
  --------- --------    -----   ----                -------------   --------    ------      -------
  23s       2s      13  {replicaset-controller }            Warning     FailedCreate    Error creating: pods "nginx-2371676037-" is forbidden: image policy webook backend denied one or more images: I don't like this images
```

### Show details of rule
In order to get information for the rule, you can use `kir get rule_name` command or display information as YAML (`kir get rule_name -o yaml`). Keep in mind that `reason` field is only available in YAML output.

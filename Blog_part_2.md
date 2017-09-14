Docker and Kubernetes
=====================

In Part 1 we walked through getting setup with Locust. We used it locally and we deployed it and our example server to Google Container Engine. In Part 2 we're going to take the same Docker image we used in Part 1 and deploy it in a distributed fashion to leverage Locust's distributed mode.

# Distributed Locust

Locust distributed mode allows you to run your locustfiles on multiple machines. You can take a look at the documentation [here](http://docs.locust.io/en/latest/running-locust-distributed.html) to learn more if you'd like. But it's a pretty straightforward setup articulated from their docs:

    To do this, you start one instance of Locust in master mode using the --master flag. This is the instance that will be running Locust’s web interface where you start the test and see live statistics. The master node doesn’t simulate any users itself. Instead you have to start one or —most likely—multiple slave Locust nodes using the --slave flag, together with the --master-host (to specify the IP/hostname of the master node).

    A common set up is to run a single master on one machine, and then run one slave instance per processor core, on the slave machines.

This design is a nice match for Kubernetes and Google Container Engine. All we need to do is create some configuration files and walk through a couple of steps and then we'll be running with as many machines as we'd like.

## Distributed Locust on Google Container Engine

In this example we'll use 7 worker nodes along with a single master node.

### Configuration Files

The configurations for our master, service definition and workers are in [kubernetes-config](/kubernetes-config). You should see the following files in that directory:

  - locust-master-controller.yaml
  - locust-master-service.yaml
  - locust-worker-controller.yaml

#### Master Controller (Replication Controller)

The master controller file (locust-master-controller.yaml) is configuring a [Kubernetes Replication Controller](https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/). From the docs:

    A ReplicationController ensures that a specified number of pod replicas are running at any one time. In other words, a ReplicationController makes sure that a pod or a homogeneous set of pods is always up and available.

The replication controller is going to be what ensures that we have the correct number of pods running for our cluster.

In our controller we've defined 1 replica that has 8 containers. They will all use the same Docker image with the same environment variables:

    - LOCUST_MODE
      - The mode for this locust instance to run in. In this case `master`
    - TARGET_HOST
      - This is mentioned in Part 1 and is the target host of the endpoint we'll be testing against.

** NOTE: Below we will walk through updating these parameters for the specific scenario we'll be running.

#### Master Service

The next file is the master service file (locust-master-service.yaml). It defines a [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/). What are services? Once again from the Kubernetes documentation:

    Kubernetes Pods are mortal. They are born and when they die, they are not resurrected. ReplicationControllers in particular create and destroy Pods dynamically (e.g. when scaling up or down or when doing rolling updates). While each Pod gets its own IP address, even those IP addresses cannot be relied upon to be stable over time. This leads to a problem: if some set of Pods (let’s call them backends) provides functionality to other Pods (let’s call them frontends) inside the Kubernetes cluster, how do those frontends find out and keep track of which backends are in that set?

    Enter Services.

    A Kubernetes Service is an abstraction which defines a logical set of Pods and a policy by which to access them - sometimes called a micro-service. The set of Pods targeted by a Service is (usually) determined by a Label Selector (see below for why you might want a Service without a selector).

In our case our service is Master Load Balancer that is exposing ports on the 8 pods it's managing.

#### Worker Controller (Replication Controller)

The last file is the worker controller (locust-worker-controller.yaml). Like our master controller the worker controller is also a ReplicationController. In this case though it's role is a worker. We have configured 10 replicas for this configuration. That number can be set to whatever you'd like but can also be modified on demand. We will show an example of that below. Also note this controller does use the same image again.

It has 3 enviornment variables it will use:

    - LOCUST_MODE:
      - Similar to the master but in this case it's `worker`
    - LOCUST_MASTER:
      - The worker needs to configure the master and we've called our master `locust-master`
    - TARGET_HOST:
      - The same target host again

Now that we know what the 3 files are we can tweak them for our use case and start the deployment process.

### Deploy Controllers and Services

#### Configure the Controller Hosts

Before deploying the `locust-master` and `locust-worker` controllers, update each to point to the location of your deployed sample web application. Set the `TARGET_HOST` environment variable found in the `locust-master-controller` and `locust-worker-controller` spec entires to the web application URL. For this example it will be the same EXTERNAL-IP as we used above. Also ensure to add the port.

    - name: TARGET_HOST
      key: TARGET_HOST
      value: http://EXTERNAL-IP:8080

#### Configure the Controller Docker Image

Now we need to tag our existing Locust Docker Image so we can deploy it to Google Image Repository. If you do not already have an image created please go to the top of this guide and follow those directions. Also ensure that you have a Google Cloud Project available to use. If you do not follow the "Before you begin" instructions from the [Google Container Engine Quick Start](https://cloud.google.com/container-engine/docs/quickstart). Once you have a Google Cloud Project setup replace PROJECT-ID throughout the rest of this guide with your Google Cloud Project Id.

    $ docker tag locust-tasks gcr.io/PROJECT-ID/locust-tasks
    $ gcloud docker -- push gcr.io/PROJECT-ID/locust-tasks

**Note:** you are not required to use the Google Container Registry. If you'd like to publish your images to the [Docker Hub](https://hub.docker.com) please refer to the steps in [Working with Docker Hub](https://docs.docker.com/userguide/dockerrepos/).

Once the Docker image has been rebuilt and uploaded to the registry you will need to edit the controllers with your new image location. Specifically, the `spec.template.spec.containers.image` field in each controller controls which Docker image to use.

If you uploaded your Docker image to the Google Container Registry:

    image: gcr.io/PROJECT-ID/locust-tasks:latest

If you uploaded your Docker image to the Docker Hub:

    image: USERNAME/locust-tasks:latest

**Note:** the image location includes the `latest` tag so that the image is pulled down every time a new Pod is launched. To use a Kubernetes-cached copy of the image, remove `:latest` from the image location.

#### Deploy Kubernetes Cluster

First create the [Google Container Engine](http://cloud.google.com/container-engine) cluster using the `gcloud` command as shown below. 

**Note:** This command defaults to creating a three node Kubernetes cluster (not counting the master) using the `n1-standard-1` machine type. Refer to the [`gcloud container clusters create`](https://cloud.google.com/sdk/gcloud/reference/container/clusters/create) documentation information on specifying a different cluster configuration.

    $ gcloud container clusters create CLUSTER-NAME

If you do not have a default Google Cloud Project Id set you can append the `--project=` argument to all of your gcloud commands like so:

    $ gcloud container clusters create CLUSTER-NAME --project=PROJECT-ID

If you do not know if you have a project set run the following command:

    $ glcoud config list

And you should see an output like so:

    [compute]
    zone = us-central1-a
    [core]
    account = email.address@somedomain.com
    project = your-google-cloud-project-id

I recommend setting both your project and compute zone. You can find the directions for both in the Google Cloud Tools Documentation.

Now let's take a look at the type of machines we can run our container nodes on. Run the following command to get a list:

    $ gcloud compute machine-types list

You can also visit the [Google Compute Machine Type Documenation](https://cloud.google.com/compute/docs/machine-types) to learn more.

In our case we're going to use some higher cpu machines as CPU is often the limiting factor when running locust. The command below uses the `n1-highcpu-8` machine type. This is a High-CPU machine type with 8 virtual CPUs and 7.20 GB of memory. Obviously the more powerful machines you use the more expensive they are. Please use what is comfortable for you. Just note that if you notice the locust system hitting failures you may need to either increase the CPU or Memory of your machines or add more machines.

Since we have 7 workers nodes with 1 controller we need to let container engine know by passing in the `--num-nodes` parameter with a value of 8 in our case. The command looks like so:

    $ gcloud container clusters create CLUSTER-NAME --machine-type=n1-highcpu-8 --num-nodes=8

Let's call our cluster here `locust-cluster` so the exact command we will run is:

    $ gcloud container clusters create locust-cluster --machine-type=n1-highcpu-8 --num-nodes=8

Now let's make the cluster our default cluster in this project by adding it to our gcloud config with the following comand:

    $ gcloud config set container/cluster locust-cluster

You can run `gcloud config list` again to confirm the following entry:

    [container]
    cluster = locust-cluster

Get credential for your cluster.

    $ gcloud container clusters get-credentials locust-cluster

Now let's look at our clusters by issuing the following command:

    $ gcloud container clusters list

Now you should see your locust-cluster. And if you created the example-cluster earlier it should also be listed.

After a few minutes, you'll have a working Kubernetes cluster with three nodes (not counting the Kubernetes master). 

Now we're going to get ready to deploy our nodes. But first we need to ensure our `kubectl` command is pointing at our cluster. If you run the following command you should see a list of contexts: 

    $ kubectl config get-clusters

Ideally you will see something like `gke_PROJECT-ID_us-central1-a_locust-cluster` and next to it will have an asterix `*` to signify that it is your default context. If it is not you will want to run the following command.

    $ kubectl config use-context gke_PROJECT-ID_ZONE_CLUSTER-NAME

Once again verify with:

    $ kubectl config get-clusters

#### Deploy locust-master

Now that `kubectl` is setup, we're going to deploy the `locust-master-controller` by issuing a create command pointed at our master controller yaml file:

    $ kubectl create -f kubernetes-config/locust-master-controller.yaml

To confirm that the Replication Controller and Pod are created, run the following:

    $ kubectl get rc

This should output something like:

    NAME            DESIRED   CURRENT   READY     AGE
    locust-master   1         1         1         41s

Now run:

    $ kubectl get pods -l name=locust,role=master

Which will output:

    NAME                  READY     STATUS    RESTARTS   AGE
    locust-master-ltg5k   1/1       Running   0          1m

If all is running we can then deploy the `locust-master-service`:

    $ kubectl create -f kubernetes-config/locust-master-service.yaml

To check the services status run:

    $ kubectl get svc

This step will expose the Pod with an internal DNS name (`locust-master`) and ports `8089`, `5557` - `5563`. As part of this step, the `type: LoadBalancer` directive in `locust-master-service.yaml` will tell Google Container Engine to create a Google Compute Engine forwarding-rule from a publicly avaialble IP address to the `locust-master` Pod. To view the newly created forwarding-rule, execute the following:

    $ gcloud compute forwarding-rules list 

#### Deploy locust-worker

Next up is our workers. We will deploy `locust-worker-controller` with the following:

    $ kubectl create -f kubernetes-config/locust-worker-controller.yaml

The `locust-worker-controller` is set to deploy 10 `locust-worker` Pods, to confirm they were deployed run the following:

    $ kubectl get pods -l name=locust,role=worker

You should see an output like so:

    NAME                  READY     STATUS    RESTARTS   AGE
    locust-worker-07m8v   1/1       Running   0          29s
    locust-worker-21f8p   1/1       Running   0          29s
    locust-worker-0j3ln   1/1       Running   0          29s
    ...

OPTIONAL: To scale the number of `locust-worker` Pods, issue a replication controller `scale` command. You can scale your pods up or down.

    $ kubectl scale --replicas=20 replicationcontrollers locust-worker

To confirm that the Pods have launched and are ready, get the list of `locust-worker` Pods:

    $ kubectl get pods -l name=locust,role=worker

**Note:** depending on the desired number of `locust-worker` Pods, the Kubernetes cluster may need to be launched with more than 3 compute engine nodes and may also need a machine type more powerful than n1-standard-1. Refer to the [`gcloud container clusters create`](https://cloud.google.com/sdk/gcloud/reference/container/clusters/create) documentation for more information.

#### Setup Firewall Rules

The final step in deploying these controllers and services is to allow traffic from your publicly accessible forwarding-rule IP address to the appropriate Container Engine instances. However your firewall rules should be added by default. However if they are not you can use this section to get them setup. If they are setup you can skip to the execution stage. To verify run the following:

    $ gcloud compute firewall-rules list

You should see some locust cluster rules listed:

    NAME                                     NETWORK  SRC_RANGES          RULES                                                                    SRC_TAGS  TARGET_TAGS
    gke-locust-cluster-19d08658-all          default  10.16.0.0/14        tcp,udp,icmp,esp,ah,sctp
    gke-locust-cluster-19d08658-ssh          default  104.155.155.130/32  tcp:22                                                                             gke-locust-cluster-19d08658-node
    gke-locust-cluster-19d08658-vms          default  10.128.0.0/9        tcp:1-65535,udp:1-65535,icmp                                                       gke-locust-cluster-19d08658-node
    k8s-fw-a022b849d81f311e79d7442010a8001e  default  0.0.0.0/0           tcp:8089,tcp:5557,tcp:5558,tcp:5559,tcp:5560,tcp:5561,tcp:5562,tcp:5563            gke-locust-cluster-19d08658-node

The target tag is the node name prefix up to `-node` and is formatted as `gke-CLUSTER-NAME-[...]-node`. For example, if your node name is `gke-mycluster-12345678-node-abcd`, the target tag would be `gke-mycluster-12345678-node`. 

If you do not see locust items listed then create the firewall rule, execute the following:

    $ gcloud compute firewall-rules create FIREWALL-RULE-NAME --allow=tcp:8089 --target-tags gke-CLUSTER-NAME-[...]-node

### Execute Tests

To execute the Locust tests, navigate to the IP address of your forwarding-rule (see above) and port `8089` and enter the number of clients to spawn and the client hatch rate then start the simulation.

You can run `kubectl get services` to view your service which will have the EXTERNAL-IP Address listed with it:

    NAME            CLUSTER-IP      EXTERNAL-IP     PORT(S)                                                                                                                   AGE
    locust-master   10.19.251.225   35.193.138.78   8089:30218/TCP,5557:30894/TCP,5558:32095/TCP,5559:32301/TCP,5560:30602/TCP,5561:30879/TCP,5562:30160/TCP,5563:31944/TCP   10m

And now you can run your tests. In our case here:

    $ open http://35.193.138.78:8089/

### Managing

For more information on managing Container Engine clusters visit the following documentation: https://kubernetes.io/docs/user-guide/managing-deployments/

### Deployment Cleanup

Once you have run your tests you will want to cleanup your locust cluster to avoid accuring costs. To teardown the workload simulation cluster, use the following steps. First, delete the Kubernetes cluster:

    $ gcloud container clusters delete locust-cluster

Next, delete the forwarding rule that forwards traffic into the cluster. To find your forwarding rule run:

    $ gcloud compute forwarding-rules list

You should see a rule that has your IP_ADDRESS. Copy the name for that entry and use it with the following command:

    $ gcloud compute forwarding-rules delete FORWARDING-RULE-NAME

Finally, delete the firewall rule that allows incoming traffic to the cluster. Similar to the forwarding rules first list the firewall rules:

    $ gcloud compute firewall-rules list

Then use the names for any firewall rules that have locust in the name with the following command:

    $ gcloud compute firewall-rules delete FIREWALL-RULE-NAME

You can use these same steps to cleanup your example app as well. You may need to update your default context first. You can switch it by listing the clusters:

    $ kubectl config get-clusters

Then take the example name and set it as default:

    $ kubectl config use-context CONTEXT-NAME

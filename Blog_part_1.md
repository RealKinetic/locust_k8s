Load Testing with Locust
========================

We believe it's critical to test our systems under load to attempt to understand the impact of system traffic. Not only the increase traffic but also the different styles of traffic. From bursty where a wall of traffic comes within a few minutes or even seconds to traffic that increases more uniformly. 

Generating load to exercise our system helps us understand how to configure our system. The different characteristics of activity can have a significant impact on how we both design and configure our systems. For example on a system like Google App Engine Google gives us the ability to configure how many idle machines (they call them instances) Google will provision beyond currently active instances. This allows us to have a buffer of machines to take on load if a wall of traffic hits. We can also configure how long a request since sits in the pending queue waiting for an instance to open up for use. In a latency sensitive environment we may configure this number to be quite low, combined with adding a buffer of idle instances when new traffic comes in it spends a short amount of time in the queue as the scheduler quickly sees that there are no available instances but thankfully instead of waiting for the instances to be spun up our traffic can be routed to the idle instances that have already been spun up. The negative of this of course is the cost of our application will go up. We may be over provisioned and now have wasted compute costs.

Load testing allows us to create artificial usage of our system that we hope to closely mimic to real life use cases. To do this we want to use our existing APIs in the same manner that our clients do. This means using realistic data and creating usage patterns similar to production. Or if we have not yet released our system we may replicate traffic based off our testing, demos or betas. Thankfully there are existing tools such as [Locust](http://locust.io/) that allow us to more easily create these test scenarios. The rest of this post walks us through installing and configuring Locust. As well as creating a simple test scenario against a simple server. In part two we will then take this same Locust setup and combine it with Google Container Engine (Google's hosted Kubernetes) to give ourselves a system that can distribute out over multiple machines and replicate significant amounts of traffic.

# Locust

## What is Locust

The Locust documentation does a good job encapsulating what Locus provides. I've quoted the high level description below but you can visit [their](http://docs.locust.io/en/latest/what-is-locust.html) documentation site here to learn more.

  Locust is an easy-to-use, distributed, user load testing tool. It is intended for load-testing web sites (or other systems) and figuring out how many concurrent users a system can handle.

  The idea is that during a test, a swarm of locusts will attack your website. The behavior of each locust (or test user if you will) is defined by you and the swarming process is monitored from a web UI in real-time. This will help you battle test and identify bottlenecks in your code before letting real users in.

  Locust is completely event-based, and therefore it’s possible to support thousands of concurrent users on a single machine. In contrast to many other event-based apps it doesn’t use callbacks. Instead it uses light-weight processes, through gevent. Each locust swarming your site is actually running inside its own process (or greenlet, to be correct). This allows you to write very expressive scenarios in Python without complicating your code with callbacks.

Now that we have a rough idea of what Locust is let's start off by installing locust.

** NOTE: If you do not wish to install Locust locally you can skip down to the Docker section towards the bottom of this guide.

## Install Locust

Locust runs within a Python environment so we will need to setup Python and it's install tools if we don't already have them. Thankfully OSX and most Linux distros come with a version of Python installed. 

To verify that you have Python installed you can open a terminal window and run the following command:

```
$ python --version
```

If you have a version of Python installed you will see a result simlar to:

```
Python 2.7.13
```

To run locust you will need either Python 2.7.x or any version of Python 3 above 3.3. If you do not have a Python runtime installed please visit the [Python site](https://www.python.org/downloads/) to download and install for your environment.

Once you have Python installed we then need to ensure with the Python tools to install packages from their repository system: [pypi](https://pypi.python.org/pypi).

Python leverages Pip to install packages locally. To check if we have pip install open a terminal window and type pip


```
$ pip
```

You will see something like below if you have pip installed:

```
Usage:
  pip <command> [options]

Commands:
  install                     Install packages.
  download                    Download packages.
  uninstall                   Uninstall packages.
  freeze                      Output installed packages in requirements format.
  list                        List installed packages.
  show                        Show information about installed packages.
  check                       Verify installed packages have compatible dependencies.
  search                      Search PyPI for packages.
  wheel                       Build wheels from your requirements.
  hash                        Compute hashes of package archives.
  completion                  A helper command used for command completion.
  help                        Show help for commands.

General Options:
...
```

If you do not have pip install please visit the [Pip installation documentation](https://pip.pypa.io/en/stable/installing/) to install.

Now that we have a working version of Python and pip we can go ahead with the Locust installation. A note for those Python users that prefer a virtual environment such as [virtualenv](https://virtualenv.pypa.io/en/stable/) you are welcome to go ahead and use a virtual environment as locust behaves fine within a virtual environment. I personally install Locust into the virtual environment for the project I am load testing but it is not a requirement.

Run this command to install locust ([alternative methods here](http://docs.locust.io/en/latest/installation.html)):

```
$ pip install locustio
```

Now that we have Locust installed we can move on to running a Locust script which requires us to have a server to hit.

## Test Server

I have a repo created that we will build out as we go. Within that repo you will find an example_server binary that was written in Go. You can either directly grab that binary here: TODO. Or you can clone the repo with the following command: TODO

Here is the Go code behind the file. It is also included in the repo if you would like to make changes or do not trust running a binary from the Internet. To build the go file run: `$ go build example_server.go`

Once you have the binary or repo pulled down go to that directly:

```
$ cd locust_k8s
```

And run the following commadn to run the server:

```
$ ./example_server
```

This server will sit on port `8080`. You can test it by visiting http://localhost:8080. You should see a page with:

  Our example home page.

There are two other endpoints exposed by this example server.

* `\login` which we will send a post request to and returns `Login. The server will output `Login Request` to stdout when this endpoint is hit.
* `\profile` which we will send a get request to and returns `Profile`. The server will output `Profile Request` to stdout when this endpoint is hit.

Now that we have an example server to hit we can create the Locust file we will use.

## Running Locust

For this example we can use the example provided by Locust in their [quick start documentation](http://docs.locust.io/en/latest/quickstart.html).

You can use the `locustfile.py` in our example repo or create said file.

Here's the full set of code that you will need to add to `locustfile.py`:

```
from locust import HttpLocust, TaskSet, task

class UserBehavior(TaskSet):
    def on_start(self):
        """ on_start is called when a Locust start before any task is scheduled """
        self.login()

    def login(self):
        self.client.post("/login", {"username":"ellen_key", "password":"education"})

    @task(2)
    def index(self):
        self.client.get("/")

    @task(1)
    def profile(self):
        self.client.get("/profile")

class WebsiteUser(HttpLocust):
    task_set = UserBehavior
    min_wait = 5000
    max_wait = 9000
```

You can learn more about what this file does in the Locust documentation and quick start walkthrough which we highly recommend you check out.

Now that we have our locustfile we can do a test.

First ensure your example server is running:

```
$ ./example_server
```

Then we run locust and give it our file.

```
locust -f locustfile.py --host=http://localhost:8080
```

We pass in the host of our example server which is running on port 8080 of our localhost.

With that now running we can open the web user interface at: http://localhost:8089

We can do a quick test by adding 1 user to simulate and 1 for a hatch rate. Then click the `Start swarming` button.

You should now see messages being sent to the stdout of your example server. In the Locust UI you will see a list of the endpoints being hit. You will see the request counts incrementing for `/` and `/profile`. There should not be failures being logged unless Locust is having issues connecting to your server.

# Deployment

We can obviously install Locust directly on any machine we'd like. Whether on bare metal, a VM, or in our case we're going to use [Docker](https://www.docker.com/) and [Google Container Engine (GKE)](https://www.docker.com/).

## Docker

We're going to build and run our Docker containers locally first so go ahead and install Docker for your environment.

## Docker Environment

We will have two containers running in our scenario. Our example server and locust instace. To support those our locust container being able to communicate with our example server we need to configure a custom docker network. Thankfully this is a simple process.

The following command will create a custom Docker network named `locustnw`

    $ docker network create --driver bridge locustnw

You can inspect this network with the following command:

    $ docker network inspect locustnw

Now that we have our network setup let's create our example server container.

### Example Server Container

To build our example server run the following:

    $ docker build examples/golang -t goexample

This will use the `Dockerfile` we've create in the `examples/golang` directory which consists of the following:

    # Start with a base Golang image
    FROM golang

    MAINTAINER Beau Lyddon <beau.lyddon@realkinetic.com>

    # Add the external tasks directory into /example
    RUN mkdir example
    ADD example_server.go example
    WORKDIR example

    # Build the example executable
    RUN go build example_server.go

    # Set script to be executable
    RUN chmod 755 example_server

    # Expose the required port (8080)
    EXPOSE 8080

    # Start Locust using LOCUS_OPTS environment variable
    ENTRYPOINT ["./example_server"] 

The `-t` argument allows us to tag our container with a name in this case we're tagging it `goexample`.

Now that we've created our container we can run it with the following:

    $ docker run -it -p=8080:8080 --name=exampleserver --network=locustnw goexample

- The `-p` argument exposes port 8080 within the container to the outside on the same port 8080. This is the port our example server is listenining on.
- The `--name` argument allows us to give a named identifier to the container. This allows us to reference this container by name as a host instead of an IP address. This will be critical when we run the locust container.
- The `--network` argument tells Docker to use our custom network for this container.

Since we exposed and mapped port 8080 you can test that our server is working by visiting http://localhost:8080.

Once you've verified that our example server container is running we can now build and run our locust container. FYI if you run locust locally earlier you can re-run the same tests again now pointing at the container version of our example server with the following `locust -f locustfile.py --host=http://localhost:8080`.

### Locust Container

Building and running our locust container is similar to our example server. First we build the container image with the following:

    $ docker build docker -t locust-tasks

This uses the `Dockerfile` in our `docker` directory. That file consists of:

    # Start with a base Python 2.7.8 image
    FROM python:2.7.13
    
    MAINTAINER Beau Lyddon <beau.lyddon@realkinetic.com>
    
    # Add the external tasks directory into /tasks
    RUN mkdir locust-tasks
    ADD locust-tasks /locust-tasks
    WORKDIR /locust-tasks
    
    # Install the required dependencies via pip
    RUN pip install -r /locust-tasks/requirements.txt
    
    # Set script to be executable
    RUN chmod 755 run.sh
    
    # Expose the required Locust ports
    EXPOSE 5557 5558 8089
    
    # Start Locust using LOCUS_OPTS environment variable
    ENTRYPOINT ["./run.sh"] 

A note. As you can see this container doesn't run locust directly but instead uses a `run.sh` file which lives in `docker/locust-tasks`. This file is important for part 2 of our tutorial when we will run locust in a distributed mode.

We will discuss quickly one import part of that file. Looking at the contents of that file:

    LOCUST="/usr/local/bin/locust"
    LOCUS_OPTS="-f /locust-tasks/locustfile.py --host=$TARGET_HOST"
    LOCUST_MODE=${LOCUST_MODE:-standalone}
    
    if [[ "$LOCUST_MODE" = "master" ]]; then
        LOCUS_OPTS="$LOCUS_OPTS --master"
    elif [[ "$LOCUST_MODE" = "worker" ]]; then
        LOCUS_OPTS="$LOCUS_OPTS --slave --master-host=$LOCUST_MASTER"
    fi
    
    echo "$LOCUST $LOCUS_OPTS"
    
    $LOCUST $LOCUS_OPTS

You see that we rely on an environment variable named `$TARGET_HOST` to be the host url passed into our locustfile. This is what will allow us to communicate across containers within our Docker network. Let's look at how we do that.

With that container built we can run it with a similar command as our dev server.

    $ docker run -it -p=8089:8089 -e "TARGET_HOST=http://exampleserver:8080" --network=locustnw locust-tasks:latest

Once again we're exposing a port but this time it's port `8089` which is the default locust port. And we pass the same network command to ensure this container also runs on our customer Docker network. However one additional argument we pass in is `-e`. This is the argument for passing in environment variables to Docker container. In this case we're passing in `http://exampleserver:8080` as the variable `TARGET_HOST`. So now we can see how the `$TARGET_HOST` environment variable in our `run.sh` script comes into play. Also we see how the custom Docker network and named containers allows us to use `exampleserver` as the host name versus attempting to find the containers IP address and passing that in. This simplifies things a great deal.

Now that we have our locust server running we can visit http://localhost:8089 in a browser on our local machine to run locust via a container hitting our dev server also running within a container.

    $ open http://localhost:8089

## Part 1 Complete

We now have a working example_server and a Locust file to run against that server. And while Locust is multi-threaded and can create a decent amount of traffic it is limited by your local resources. Even pushing it to a powerful hosted machine is going to hit limitations. The true power of Locust comes in its ability to distribute out over multiple machines. However creating clustered enviornments can be a bit of a pain. In part two we'll walkthrough leveraging Google Compute Enging (Kubernetes) and Locust's distributed mode to give us a maintainable distributed environment to run our load tests from.

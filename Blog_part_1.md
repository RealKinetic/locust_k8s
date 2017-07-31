Load Testing with Locust and Kubernetes - Part 1
=================================================

We believe it's critical to test our systems under load to attempt to understand the impact of system traffic. Not only the increase traffic but also the different styles of traffic. From bursty where a wall of traffic comes within a few minutes or even seconds. To traffic that increases more uniformly. 

Generating load to exercise our system helps us understand how to configure our system. Especially in the elastic environments provider by many cloud vendors. Even though these systems are elastic we must design to take advantage of this elasticity. The different characteristics of activity can have a significant impact on how we both design and configure the system. For example on a system like Google App Engine they give us the ability to configure how many idle machines (they call them instances) Google will provision beyond currently active instances. This allows us to have a buffer of machines to take on load if a wall of traffic hits. We can also configure how long a request since sits in the pending queue waiting for an instance to open up for use. In a latency sensitive environment we may configure this number to be quite low, combined with adding a buffer of idle instances when new traffic comes in it spends a short amount of time in the queue as the scheduler quickly sees that there are no available instances but thankfully instead of waiting for the instances to be spun up our traffic can be routed to the idle instances that have already been spun up. The negative of this of course is the cost of our application will go up. We may be over provisioned and now have wasted compute costs.

Load testing allows us to create artificial usage of our system that we hope to closely mimic to real life use cases. To do this we want to use our existing APIs in the same manner that our clients do. This means using realistic data and creating usage patterns similar to production. Or if we have not yet released our system we may replicate traffic based off our testing, demos or betas. Thankfully there are existing tools such as [Locust](http://locust.io/) that allow us to more easily create these test scenarios. The rest of this post walks us through installing and configuring Locust. As well as creating a simple test scenario against a simple server. In part two we will then take this same Locust setup and combine it with Kubernetes to give ourselves a system that can distribute out over multiple machines and replicate significant amounts of traffic.

### What is Locust

The Locust documentation does a good job encapsulating what Locus provides. I've quoted the high level description below but you can visit [their](http://docs.locust.io/en/latest/what-is-locust.html) documentation site here to learn more.

  Locust is an easy-to-use, distributed, user load testing tool. It is intended for load-testing web sites (or other systems) and figuring out how many concurrent users a system can handle.

  The idea is that during a test, a swarm of locusts will attack your website. The behavior of each locust (or test user if you will) is defined by you and the swarming process is monitored from a web UI in real-time. This will help you battle test and identify bottlenecks in your code before letting real users in.

  Locust is completely event-based, and therefore it’s possible to support thousands of concurrent users on a single machine. In contrast to many other event-based apps it doesn’t use callbacks. Instead it uses light-weight processes, through gevent. Each locust swarming your site is actually running inside its own process (or greenlet, to be correct). This allows you to write very expressive scenarios in Python without complicating your code with callbacks.

### Install Locust

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

### Test Server

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

### Running Locust

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

### Part 1 Complete

We now have a working example_server and a Locust file to run against that server. And while Locust is multi-threaded and can create a decent amount of traffic it is limited by your local resources. Even pushing it to a powerful hosted machine is going to hit limitations. The true power of Locust comes in its ability to distribute out over multiple machines. However creating clustered enviornments can be a bit of a pain. In part two we'll walkthrough leveraging Kubernetes to give us a maintainable distributed environment to run our load tests from.

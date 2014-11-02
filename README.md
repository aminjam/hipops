#hipops
**Docker Orchestration Configuration**  

hipops is an agentless orchestration configuration tool for servers running [Docker](https://github.com/docker/docker) containers, whether it's local or remote.
It's a JSON-based configuration of your environments. I created hipops as a wrapper around the popular orchestration tool, [ansible](https://github.com/ansible/ansible), for ease of use.
Ansible Playbooks are powerful for managing configurations and deployments for the remote machines or local VMs. When it comes to docker deployments, there is a need for a configurable orchestration of containers. `hipops` will allow you to orchestrate the cluster of your servers (`CoreOS`, `ubuntu`, etc...) with a configuration template locally in VMs and then remotely to AWS or other providers. `hipops` works with [CoreOS](https://CoreOS.com/), [ubuntu](http://www.ubuntu.com/), and other linux servers.

##Concept
The idea behind `hipops` configuration is to define a series of `apps` and re-use them in `playbooks`, so you just focus on the orchestration of your containers, whether across different physical hosts or the same host. Here is a sample configuration for `hipops`:
```
{
  "id": "demo",
  "description": "my demo",
  "env": "dev",
  "dest": "/data",
  "oses": [{
    "user": "ubuntu"
  }],
  "apps": [{
    "name": "mongo",
    "type": "db",
    "image": "aminjam/mongodb:latest",
    "ports": [27017]
  }, {
    "name": "backend-api",
    "type": "nodejs",
    "host": "backend-api-host.com",
    "repository": {
      "branch": "master",
      "sshUrl": "github.com/aminjam/hipops-SAMOMY-backend.git"
    },
    "image": "aminjam/nodejs:latest",
    "ports": [3001]
  }],
  "playbooks": [{
    "inventory": "tag_App-Role_DEMO",
    "apps": ["{{index .Apps 0}}"],
    "actions": [{
      "params": "-v {{.App.Dest}}:/home/app -d {{.App.Image}} /run.sh"
    }]
  }, {
    "inventory": "tag_App-Role_DEMO",
    "apps": ["{{index .Apps 1}}"],
    "state": "deploying",
    "actions": [{
      "params": "-v {{.App.Dest}}:/home/app -e NODE_ENV=development --link {{(index .Apps 0).Name}}:mongo -d {{.App.Image}} /run.sh"
    }]
  }]
}
```
I am defining two apps: `mongo` and `backend-api`, and then I define the first `playbook` to run `{{index .Apps 0}}` which in this case is `mongo` and then the second `playbook` to run `{{index .Apps 1}}` which is `backend-api`, and since we are only focusing on configuring the containers, we are tapping into ansible's power of running the inventories against `tag_App-Role_DEMO` machines (machines tagged with App-Role=DEMO)

##Next steps
- [Getting Started Guide](https://github.com/aminjam/hipops/wiki/Getting-Started)
- [JSONish Configuration explained](https://github.com/aminjam/hipops/wiki/JSONish-Configuration)
- Checkout `devops` folder for sample scenarios for:
  - **SAMOMY-dev**: (S)sailsJS-backend + (A)angular-frontend + (MO)mongodb + (MY)mysql on a single host (CoreOS)
  - **SAMOMY-prod**: (S)sailsJS-backend + (A)angular-frontend + (MO)mongodb + (MY)mysql linked together on three different hosts (2 ubuntu, 1 CoreOS)
  - **SD-CR**: (S)service (D)discovery with (C)consul + (R)registrator on all of your servers (ubuntu)
  - **ELKF-prod**: (E)elasticsearch + (L)logstash + (K)kabana + (F)logstash-forwarder for aggregating the logs across all containers (ubuntu)
  - **CBnodejs-demo**: (CB)couchbase-server + nodejs demo app (CoreOS)

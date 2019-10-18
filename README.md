# terraform-gcp-yugabyte
A Terraform module to deploy and run YugaByte on Google Cloud.

## Config
* First create a terraform file and add the yugabyte terraform module to your file 
  ```
  module "yugabyte-db-cluster" {
  source = "github.com/YugaByte/terraform-gcp-yugabyte.git"

  # Your GCP project id 
  project_id = "<YOUR-GCP-PROJECT-ID>"

  # Your GCP credentials file path  
  credentials = "<PATH-OF-YOUR-GCP-CREDENTIAL-FILE>"

  # The name of the cluster to be created.
  cluster_name = "test-yugabyte"

  # User name for ssh connection
  ssh_user = "SSH_USER_NAME_HERE"

  # The region name where the nodes should be spawned.
  region_name = "YOUR VPC REGION"

  # Replication factor.
  replication_factor = "3"

  # The number of nodes in the cluster, this cannot be lower than the replication factor.
  node_count = "3"
  }
  ```
  Note:- You can get credentials file by following steps given [here](https://cloud.google.com/docs/authentication/getting-started)


## Usage

Init terraform first if you have not already done so.

```
$ terraform init
```

To check what changes are going to happen in the environment run the following 

```
$ terraform plan
```


Now run the following to create the instances and bring up the cluster.

```
$ terraform apply
```

Once the cluster is created, you can go to the URL `http://<node ip or dns name>:7000` to view the UI. You can find the node's ip or dns by running the following:

```
terraform state show google_compute_instance.yugabyte_node[0]
```

You can access the cluster UI by going to any of the following URLs.

You can check the state of the nodes at any point by running the following command.

```
$ terraform show
```

To destroy what we just created, you can run the following command.

```
$ terraform destroy
```
`Note:- To make any changes in the created cluster you will need the terraform state files. So don't delete state files of Terraform.`

## Test 

### Configurations

#### Prerequisites

- [Terraform **(~> 0.12.5)**](https://www.terraform.io/downloads.html)
- [Golang **(~> 1.12.10)**](https://golang.org/dl/)
- [dep **(~> 0.5.4)**](https://github.com/golang/dep)

#### Environment setup

* First install `dep` dependency management tool for Go.
    ```sh
    $ curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
    ```  
* Change your working directory to the `test` folder.
* Run `dep` command to get required modules
    ```sh
    $ dep ensure
    ```

#### Run test

Then simply run it in the local shell:

```sh
$ go test -v -timeout 15m  yugabyte_test.go
```
* Note that go has a default test timeout of 10 minutes. With infrastructure testing, your tests will surpass the 10 minutes very easily. To extend the timeout, you can pass in the -timeout option, which takes a go duration string (e.g 10m for 10 minutes or 1h for 1 hour). In the above command, we use the -timeout option to override to a 90 minute timeout.
* When you hit the timeout, Go automatically exits the test, skipping all cleanup routines. This is problematic for infrastructure testing because it will skip your deferred infrastructure cleanup steps (i.e terraform destroy), leaving behind the infrastructure that was spun up. So it is important to use a longer timeout every time you run the tests.


# terraform-gcp-yugabyte
A Terraform module to deploy and run YugaByte on Google Cloud.

## Config
* First create a terraform file with provider details 
  ```
  provider "google" 
  { 
    # Provide your GCP Creadentilals 
    credentials = "${file("yugabyte-pcf-bc8114281026.json")}"

    # The name of your GCP project 
    project = "yugabyte-pcf"
  }
  ```
  Note :- You can get credentials file by following steps given [here](https://cloud.google.com/docs/authentication/getting-started)

* Now add the yugabyte terraform module to your file 
  ```
  module "yugabyte-db-cluster" {
  source = "github.com/YugaByte/terraform-gcp-yugabyte.git"

  # The name of the cluster to be created.
  cluster_name = "test-yugabyte"

   # key pair.
  ssh_private_key = "SSH_PRIVATE_KEY_HERE"
  ssh_public_key = "SSH_PUBLIC_KEY_HERE"
  ssh_user = "SSH_USER_NAME_HERE"

  # The region name where the nodes should be spawned.
  region_name = "YOUR VPC REGION"

  # Replication factor.
  replication_factor = "3"

  # The number of nodes in the cluster, this cannot be lower than the replication factor.
  node_count = "3"
  }
  ```


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

resource "google_cloud_run_service" "default" {
    name = "chat-app"
    location = "us-central1"
    provider = google-beta
    autogenerate_revision_name = true
    
    template {
        spec {
            containers {
                image = "gcr.io/chat-app-338321/chatappserver@sha256:6f49befc253ffbe977e4dd2ccab5e35722bb2ae8abe0e9b49653221df96d2800"
            }
        }

        metadata {
            annotations = {
                    "run.googleapis.com/vpc-access-connector" = google_vpc_access_connector.connector.name
                    "run.googleapis.com/vpc-access-egress" = "all"
                    "run.googleapis.com/client-name" = "terraform"
            }
        }
    }

    traffic {
        percent = 100
    }
}

data "google_iam_policy" "noauth" {
   binding {
     role = "roles/run.invoker"
     members = ["allUsers"]
   }
 }

 resource "google_cloud_run_service_iam_policy" "noauth" {
   location    = google_cloud_run_service.default.location
   project     = google_cloud_run_service.default.project
   service     = google_cloud_run_service.default.name

   policy_data = data.google_iam_policy.noauth.policy_data
}

resource "google_redis_instance" "default" {
    name =  "chat-app-pubsub"
    tier = "BASIC"
    memory_size_gb = 1

    location_id ="us-central1-a"

   # reserved_ip_range = "192.168.0.0/29"

    labels = {
      my_key    = "my_val"
      other_key = "other_val"
    }
}

resource "google_compute_network" "default" {
  name                    = "cloudrun-network"
  provider                = google-beta
  auto_create_subnetworks = false
}

# VPC access connector
resource "google_vpc_access_connector" "connector" {
  name          = "vpcconn"
  provider      = google-beta
  region        = "us-central1"
  ip_cidr_range = "10.8.0.0/28"
  network       = google_compute_network.default.name
  depends_on    = [google_project_service.vpcaccess_api]
}


module "test-vpc-module" {
  source       = "terraform-google-modules/network/google"
  version      = "~> 3.3.0"
  project_id   = var.project 
  network_name = "my-serverless-network"
  mtu          = 1460

  subnets = [
    {
      subnet_name   = "serverless-subnet"
      subnet_ip     = "10.10.10.0/28"
      subnet_region = "us-central1"
    }
  ]
}

module "serverless-connector" {
  source     = "terraform-google-modules/network/google//modules/vpc-serverless-connector-beta"
  project_id = var.project
  vpc_connectors = [{
    name        = "central-serverless"
    region      = "us-central1"
    subnet_name = module.test-vpc-module.subnets["us-central1/serverless-subnet"].name
    machine_type  = "e2-standard-4"
    min_instances = 2
    max_instances = 7
    }
      , {
        name          = "central-serverless2"
        region        = "us-central1"
        network       = module.test-vpc-module.network_name
        ip_cidr_range = "10.10.11.0/28"
        subnet_name   = null
        machine_type  = "e2-standard-4"
        min_instances = 2
      max_instances = 7 }
  ]
  depends_on = [
    google_project_service.vpcaccess_api
  ]
}



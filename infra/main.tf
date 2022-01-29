terraform {
  required_providers {
      google = {
          source = "hashicorp/google"
          version = "~> 3.5"
      }

  }
}
provider "google" {
    project = var.project

    credentials = file("/home/midepeter/Desktop/Go/chat-app/chat-app-338321-d86b6075c3de.json")
    region = "us-central1"
}


provider "google-beta" {

    credentials = file("/home/midepeter/Desktop/Go/chat-app/chat-app-338321-d86b6075c3de.json")

    project = var.project

    region = "us-central1"
}
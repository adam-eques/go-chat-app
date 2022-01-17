resource "google_project_service" "default" {
    service = "run.googleapis.com"
    disable_on_destroy = true
}

resource "google_project_service" "vpcaccess_api" {
    service = "vpcaccess.googleapis.com"
    provider = google-beta
    disable_on_destroy = false
}
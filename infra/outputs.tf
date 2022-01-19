output "service_url" {
  value = google_cloud_run_service.default.status[0].url
}

output "host" {
 description = "The IP address of the instance."
 value = "${google_redis_instance.default.host}"
}
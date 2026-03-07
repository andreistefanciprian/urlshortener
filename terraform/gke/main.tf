locals {
  labels = merge({
    managed_by = "terraform"
    env        = var.project_name
  }, var.labels)
}

# GKE cluster
resource "google_project_service" "container" {
  service                    = "container.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_project_service" "compute" {
  service                    = "compute.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# Wait for APIs to propagate after enablement
resource "time_sleep" "wait_for_apis" {
  create_duration = "30s"
  depends_on = [
    google_project_service.container,
    google_project_service.compute,
  ]
}

resource "google_container_cluster" "primary" {
  provider           = google-beta
  name               = var.project_name
  location           = var.gcp_region
  network            = data.terraform_remote_state.networking.outputs.vpc_name
  subnetwork         = data.terraform_remote_state.networking.outputs.subnet_name
  logging_service    = "logging.googleapis.com/kubernetes"    # Lets use Stackdriver
  monitoring_service = "monitoring.googleapis.com/kubernetes" # Lets use Stackdriver

  enable_shielded_nodes       = true         # https://cloud.google.com/kubernetes-engine/docs/how-to/shielded-gke-nodes
  enable_intranode_visibility = true         # https://cloud.google.com/kubernetes-engine/docs/how-to/intranode-visibility
  networking_mode             = "VPC_NATIVE" # Required for private service networking to services

  deletion_protection = false

  private_cluster_config {
    enable_private_endpoint = true
    enable_private_nodes    = true
    master_ipv4_cidr_block  = var.gke_master_cidr # 10.100.48.0/28
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = data.terraform_remote_state.networking.outputs.subnet_ip_cidr_range # 10.100.0.0/24
      display_name = "VPC subnet"
    }
  }

  # https://cloud.google.com/kubernetes-engine/docs/how-to/alias-ips
  ip_allocation_policy {
    cluster_secondary_range_name  = "gke-pods"     # 10.100.16.0/20
    services_secondary_range_name = "gke-services" # 10.100.32.0/24
  }

  master_auth {
    client_certificate_config {
      issue_client_certificate = false
    }
  }

  # recommended way to safely access Google Cloud services from GKE applications.
  workload_identity_config {
    workload_pool = "${var.gcp_project}.svc.id.goog" # https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
  }

  maintenance_policy {
    daily_maintenance_window {
      start_time = var.maintenance_window
    }
  }

  vertical_pod_autoscaling {
    enabled = true
  }

  release_channel {
    channel = "REGULAR"
  }

  # When use_default_node_pool=false: remove the default pool immediately (replaced by separately managed pool).
  # When use_default_node_pool=true: keep the default pool and configure it with the same node settings.
  remove_default_node_pool = !var.use_default_node_pool

  initial_node_count = 1

  dynamic "node_config" {
    for_each = var.use_default_node_pool ? [1] : []
    content {
      machine_type = var.node_type
      spot         = true
      disk_size_gb = var.node_disk_size
      disk_type    = var.node_disk_type

      service_account = google_service_account.cluster.email
      oauth_scopes = [
        "https://www.googleapis.com/auth/cloud-platform"
      ]

      labels = local.labels

      tags = ["gke-node", var.project_name]
    }
  }

  addons_config {

    http_load_balancing {
      disabled = false
    }

    network_policy_config {
      disabled = false
    }

    gce_persistent_disk_csi_driver_config {
      enabled = true
    }
  }

  network_policy {
    enabled  = true
    provider = "CALICO"
  }

  resource_labels = merge(local.labels, {
    mesh_id = "proj-${data.google_project.project.number}"
  })

  timeouts {
    create = "30m"
    update = "40m"
  }

  depends_on = [
    google_service_account.cluster,
    time_sleep.wait_for_apis,
  ]
}

# Separately managed node pool (not used when use_default_node_pool=true)
resource "google_container_node_pool" "primary_nodes" {
  count    = var.use_default_node_pool ? 0 : 1
  name     = var.project_name
  location = var.gcp_region
  cluster  = google_container_cluster.primary.name

  autoscaling {
    total_min_node_count = 1
    total_max_node_count = 2
    location_policy      = "ANY"
  }

  node_config {
    machine_type = var.node_type
    spot         = true
    disk_size_gb = var.node_disk_size
    disk_type    = var.node_disk_type

    service_account = google_service_account.cluster.email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = local.labels

    tags = ["gke-node", var.project_name]
  }

  lifecycle {
    ignore_changes = [
      node_config[0].resource_labels,
      node_config[0].kubelet_config,
    ]
  }
}
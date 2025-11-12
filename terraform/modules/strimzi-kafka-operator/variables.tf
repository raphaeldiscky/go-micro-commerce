# Strimzi Kafka Operator Module Variables

variable "namespace" {
  description = "Kubernetes namespace for the operator"
  type        = string
  default     = "kafka-system"
}

variable "create_namespace" {
  description = "Create the namespace if it doesn't exist"
  type        = bool
  default     = true
}

variable "chart_version" {
  description = "Helm chart version for Strimzi Kafka operator"
  type        = string
  default     = "0.44.0"
}

variable "watch_namespaces" {
  description = "Namespaces to watch for Kafka resources (empty = all namespaces)"
  type        = list(string)
  default     = []
}

variable "log_level" {
  description = "Log level for the operator"
  type        = string
  default     = "INFO"
}

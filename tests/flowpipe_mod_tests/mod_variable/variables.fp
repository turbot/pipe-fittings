
variable "mandatory_tag_keys" {
  type        = list(string)
  description = "A list of mandatory tag keys to check for (case sensitive)."
  default     = ["Environment", "Owner"]
}


variable "var_number" {
  type        = number
  default = 42
}

variable "var_map" {
    type = map(string)
    default = {
        key1 = "value1"
        key2 = "value2"
    }
}
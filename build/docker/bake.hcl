variable "IMAGE_NAME" {
  default = ""
}

variable "DOCKERFILE_PATH" {
  default = "Dockerfile"
}

variable "CI_COMMIT_TAG" {
  default = ""
}

variable "CI_COMMIT_REF_SLUG" {
  default = ""
}

variable "CI_COMMIT_SHORT_SHA" {
  default = ""
}

variable "IMAGE_TAG" {
  default = notequal("", CI_COMMIT_TAG) ? "${CI_COMMIT_TAG}" : "${CI_COMMIT_REF_SLUG}-${CI_COMMIT_SHORT_SHA}"
}

variable "BUILD_TAG" {
  default = notequal("-", IMAGE_TAG) ? "${IMAGE_NAME}:${IMAGE_TAG}" : "${IMAGE_NAME}:latest"
}

group "default" {
  targets = ["go-tools"]
}

target "go-tools"{
  dockerfile = "${DOCKERFILE_PATH}"
  context    = "."
  platforms  = ["linux/amd64", "linux/arm64"]
    tags       = [
        "${BUILD_TAG}"
    ]
}

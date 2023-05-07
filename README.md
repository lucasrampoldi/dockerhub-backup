This program retrieves image tags for a repository, downloads image, compresses and uploads to S3.

For example, you want to migrate your current repository to another repository and you don't want to lose the images.

#### COMPILE
$go build main.go

#### RUN
$./main repo_name organization_name docker_username docker_password bucket_name

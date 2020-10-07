package main

import (
	"./package"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/docker/docker/api/types"
	. "github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)
const defaultDockerAPIVersion = "v1.40"


func UploadToS3(bucket string,repository string, filename string) {
	time := time.Now().Local().Format("2006.01.02-15:04:05")
	file, err := os.Open("/tmp/"+filename)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:    aws.String(repository),
	})

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key: aws.String(repository+"/"+time+"_"+filename),
		Body: file,
	})
	if err != nil {
		exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
func GetTagListFromDockerAPI(username string, password string,org string,repository string) _package.Tags{
		//do login
	payloadBytes, err := json.Marshal(_package.Payload{Username: username, Password:password })
	request, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/",  bytes.NewReader(payloadBytes))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()


	//get token
	bodyBytes, err := ioutil.ReadAll(response.Body)
	var token _package.Token
	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		log.Fatal(err)
	}


	//lists tags for a repository
	request, err = http.NewRequest("GET", os.ExpandEnv("https://hub.docker.com/v2/repositories/"+org+"/"+repository+"/tags/?page_size=10000"), nil)
	request.Header.Set("Authorization", os.ExpandEnv("JWT "+token.Token))
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	bodyBytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	var tags _package.Tags
	err = json.Unmarshal(bodyBytes, &tags)

	return tags
}

func DoBackupFromDockerCli(username string,password string,org string, repository string,tag _package.ResultTags,bucket string)  {

	//pull image by tag
	client, err := NewClientWithOpts(WithVersion(defaultDockerAPIVersion))
	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	out, err := client.ImagePull(context.Background(),org+"/"+repository+":"+tag.Name,types.ImagePullOptions{RegistryAuth: base64.URLEncoding.EncodeToString(encodedJSON)})
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	//save image
	var imageSlice []string
	imageSlice = append(imageSlice,org+"/"+ repository+":"+tag.Name)
	image,_ := client.ImageSave(context.Background(),imageSlice)
	buf := new(bytes.Buffer)
	buf.ReadFrom(image)

	//create file and compress image
	writer, err := os.Create("/tmp/"+repository+":"+tag.Name+"-"+tag.LastUpdated+".gz")
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = repository+":"+tag.Name
	defer archiver.Close()
	io.Copy(archiver, buf)

	//remove image
	_,err = client.ImageRemove(context.Background(),org+"/"+ repository+":"+tag.Name,types.ImageRemoveOptions{})
	if err != nil {
		log.Fatal(err)
	}

	//upload to S3
	UploadToS3(bucket,repository,repository+":"+tag.Name+"-"+tag.LastUpdated+".gz")
}
func main() {
	if len(os.Args) != 6 {
		exitErrorf("The repository name, organization name, Docker Username, Docker Password and Bucket name are required.(5 args)\n")
	}
	fmt.Println("Starting...")

	repository := os.Args[1]
	org := os.Args[2]
	username:= os.Args[3]
	password:= os.Args[4]
	bucket:= os.Args[5]


	tags:=GetTagListFromDockerAPI(username,password,org,repository)
	fmt.Println(len(tags.Results),"tags found...")
	for index := 0 ;index < len(tags.Results);index++{
		DoBackupFromDockerCli(username,password,org,repository,tags.Results[index],bucket)

	}
	fmt.Println("Finished.")

}
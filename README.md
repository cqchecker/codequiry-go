# Go SDK for Codequiry API 

Codequiry is a commercial grade plagiarism and similarity detection software for source code files. Submissions are checked with billions of sources on the web as well as checked locally against provided submissions. This is a NodeJS example application for the API to check code plagiarism and similarity.

The API allows us to run multiple different tests on source code files: 
1. Peer Check - Given a group of submissions as individual zip files, all lines of code are compared to each other and relative similarity scores are computed, as well as matched snippets. 
2. Database Check - Checks submissions against popular repositories and public sources of code.
3. Web Check - Does a full check of code with over 2 billion public sources on the web. 

Checks return us tons of data such as similarity scores, individual file scores, cluster graphs, similarity histograms, highlights results, matched snippets, percentage plagiarised and similar, and a ton more... 

Main Website: 
https://codequiry.com

Full API Docs:
https://codequiry.com/usage/api

## Installation
```
go get -u "github.com/cqchecker/codequiry-go"
```
#### Import it with
```
import "github.com/cqchecker/codequiry-go"
```

#### Setting your API Key
```
var cq = Codequiry{ApiKey: "YOUR_API_KEY"}
```
## Usage
#### Getting account information
```go
account, err := cq.Account()
fmt.Printf("%+v\n %+v\n",account, err)
```
#### Getting checks
```go
checks, err := cq.Checks()
fmt.Printf("%+v\n %+v\n",checks, err)
```
#### Creating checks (specify name and programming language)
Examples: java, c-cpp, python, csharp, txt
```go
check, err := cq.CreateCheck("CheckNameHere", "java")
fmt.Printf("%+v\n %+v\n",check, err)

```
#### Uploading to a check (specify check_id and file (must be a zip file)) 
```go
res, err := cq.UploadFile("CHECK_ID", "./test.zip")
fmt.Printf("%+v\n %+v\n",res, err)
```
#### Starting a check (specify check_id and if running database check or web check) 
```go
status, err := cq.startCheck("CHECK_ID")
fmt.Printf("%+v\n %+v\n",status, err)
```
#### Getting a check information/status
```go
check, err := cq.getCheck("CHECK_ID")
fmt.Printf("%+v\n %+v\n",check, err)
```
#### Getting results overview
```go
overview, err := cq.getOverview("CHECK_ID")
fmt.Printf("%+v\n %+v\n",overview, err)
```
#### Getting specific results of a submission
```go
results, err := cq.getResults("CHECK_ID", "SUBMISSION_ID")
fmt.Printf("%+v\n %+v\n",results, err)
```
## Realtime checking progress - SocketIO
This is an example of the listener, you can call this after getting a check status or after starting a check (both will reutrn a job ID, which you can listen to). Here we will listen to specific CHECK_ID.
```go
callback := func(data string) {
    print("Update: " + data)
}

cq.checkListen(1, callback)
```
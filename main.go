package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/pkg/errors"
    "github.com/zhouhui8915/go-socket.io-client"
    "io/ioutil"
    "mime/multipart"
    "net/http"
    "os"
    "reflect"
    "regexp"
    "strings"
    "time"
)

const apiBaseUrl = "https://codequiry.com/api/v1/"
const apiUploadUrl = "https://codequiry.com/api/v1/check/upload"
const socketsBaseUrl = "https://api.codequiry.com/"

var (
    ErrServer                  = errors.New("unexpected error when making a request to the API")
    ErrSocketConnection        = errors.New("there was an error when trying to establish a socket connection")
    ErrJobCheck                = errors.New("error when trying to check the job status")
)

type Codequiry struct {
    ApiKey string
}

type Account struct {
    User                    string
    Email                   string
    PeerChecksRemaining     string                 	`json:"peer_checks_remaining"`
    ProChecksRemaining      int                    	`json:"pro_checks_remaining"`
    Submissions             int
}

type Check struct {
    Id                      int
    Name                    string
    CreatedAt               date                	`json:"created_at"`
    UpdatedAt               date                	`json:"updated_at"`
    StatusId                int                     `json:"status_id"`
    JobId                   int                     `json:"job_id"`
}

type CheckStatusInfo struct {
    Check                   Check
    Status                  string
    DBCheck                 bool
    WebCheck                bool
    SubmissionCount         int                     `json:"submission_count"`
    DashUrl                 string
}

type Submission struct {
    Id                      int
    Filename                string
    StatusId                int                    	`json:"status_id"`
    CreatedAt               date                	`json:"created_at"`
    UpdatedAt               date                	`json:"updated_at"`
    Result1                 float32
    Result2                 float32
    Result3                 float32
    TotalResult             float32
    SubmissionResults       []SubmissionResult    	`json:"submission_results"`
}

type SubmissionResult struct {
    Id                      int
    SubmissionId            int                     `json:"submission_id"`
    SubmissionIdCompared    int                    	`json:"submission_id_compared"`
    Score                   float32
    CreatedAt               date                	`json:"created_at"`
    UpdatedAt               date                	`json:"updated_at"`
}

type Overview struct {
    OverviewURL             string
    Submissions             []Submission
    bardata                 []interface{}
}

type RelatedFile struct {
    ID                      int
    SubmissionID            int                 	`json:"submission_id"`
    Filedir                 string
    Content                 string
    CreatedAt               date                 	`json:"created_at"`
    UpdatedAt               date                 	`json:"updated_at"`
    LanguageID              int                 	`json:"language_id"`
}

type PeerMatch struct {
    ID                      int
    SubmissionID            int                 	`json:"submission_id"`
    SubmissionIDMatched     int                 	`json:"submission_id_matched"`
    Similarity              string
    MatchedSimilarity       string              	`json:"matched_similarity"`
    File                    string
    FileMatched             string              	`json:"file_matched"`
    LineStart               int                 	`json:"line_start"`
    LineEnd                 int                 	`json:"line_end"`
    Tokens                  int
    CreatedAt               interface{}         	`json:"created_at"`
    UpdatedAt               interface{}         	`json:"updated_at"`
    LineMatchedStart        int                 	`json:"line_matched_start"`
    LineMatchedEnd          int                 	`json:"line_matched_end"`
    MatchType               int                 	`json:"match_type"`
}

type SubmissionResults struct {
    Submission              Submission
    Avg                     float64
    Max                     string
    Min                     string
    PeerMatches             []PeerMatch
    OtherMatches            []interface{}			`json:"other_matches"`
    RelatedSubmissions      []Submission        	`json:"related_submissions"`
    RelatedFiles            []RelatedFile         	`json:"related_files"`
}

type AssignmentStatus struct {
    ID                    	int
    Status                  string
    Icon                    interface{}
    Color                   string
    CreatedAt               date                  	`json:"created_at"`
    UpdatedAt               date                  	`json:"updated_at"`
}

type UploadData struct {
    ID                    	int
    Filename                string
    StatusID                int                  	`json:"status_id"`
    CreatedAt               string               	`json:"created_at"`
    UpdatedAt               string               	`json:"updated_at"`
    Result1                 string
    Result2                 string
    Result3                 string
    TotalResult             string               	`json:"total_result"`
    ModifyUpdatedAt         string               	`json:"modify_updated_at"`
    AssignmentStatuses      []AssignmentStatus
    File                    string
    SubmissionCount         int                 	`json:"submission_count"`
    Check                   Check
}

type APIError struct {
    msg                     string
}

func (e* APIError) Error() string {
    return fmt.Sprintf(e.msg)
}

var errorRegexp = regexp.MustCompile(`^{"error"\s*:\s*"(?P<msg>.*)"}$`)

type date struct {
    time.Time
}

func (c Codequiry) GetBaseHeaders() http.Header {
    header := http.Header{}
    header.Set("Content-Type", "application/json")
    header.Set("apikey", c.ApiKey)
    return header
}

func (c Codequiry) Account() (Account, error) {
    jsonStr, err := c.post("account", nil, nil)
    account := Account{}
    err = unmarshal(jsonStr, &account)

    return account, err
}

func (c Codequiry) Checks() ([]Check, error) {
    jsonStr, err := c.post("checks", "", nil)
    var checks []Check
    if err != nil {
        return checks, err
    }
    err = unmarshal(jsonStr, &checks)

    return checks, err
}

func (c Codequiry) CreateCheck(checkName string, lang string) ([]Check, error) {
    params := make(map[string]interface{})
    params["name"] = checkName
    params["language"] = lang

    jsonStr, err := c.post("check/create", params, nil)
    var checks []Check
    err = unmarshal(jsonStr, &checks)

    return checks, err
}

func (c Codequiry) CheckListen(jobId int, callback func(string)) error {
    opts := &socketio_client.Options{
        Transport: "websocket",
        Query:     make(map[string]string),
    }

    client, err := socketio_client.NewClient(socketsBaseUrl, opts)
    if err != nil {
        return ErrSocketConnection
    }

    _ = client.On("connection", func() error {
        params := make(map[string]interface{})
        params["jobid"] = jobId
        err = client.Emit("job-check", params)
        if err != nil {
            return ErrJobCheck
        }

        return err
    })

    _ = client.On("job-status", func(msg string) {
        callback(msg)
        var response struct {
            error   int
            percent int
        }
        _ = json.Unmarshal([]byte(msg), &response)
        if response.error == 1 || response.percent == 100 {
            return
        }
    })

    return nil
}

func (c Codequiry) StartCheck(checkId string) (CheckStatusInfo, error) {
    params := make(map[string]interface{})
    params["check_id"] = checkId

    jsonStr, err := c.post("check/start", params, nil)
    var checkStatus CheckStatusInfo
    err = unmarshal(jsonStr, &checkStatus)

    return checkStatus, err
}

func (c Codequiry) GetCheck(checkId string) (Check, error) {
    params := make(map[string]interface{})
    params["check_id"] = checkId

    jsonStr, err := c.post("check/get", params, nil)
    var check Check
    err = unmarshal(jsonStr, &check)

    return check, err
}

func (c Codequiry) GetOverview(checkId string) (Overview, error) {
    params := make(map[string]interface{})
    params["check_id"] = checkId

    jsonStr, err := c.post("check/overview", params, nil)
    var overview Overview
    err = unmarshal(jsonStr, &overview)

    return overview, err
}

func (c Codequiry) GetResults(checkId string, sid string) (SubmissionResults, error) {
    params := make(map[string]interface{})
    params["check_id"] = checkId
    params["submission_id"] = sid

    jsonStr, err := c.post("check/results", params, nil)
    var results SubmissionResults
    err = unmarshal(jsonStr, &results)

    return results, err
}

func (c Codequiry) UploadFile(checkId string, filePath string) ([]UploadData, error) {
    file, _ := os.Open(filePath)
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    fileContents, _ := ioutil.ReadFile(filePath)
    _ = writer.WriteField("file", string(fileContents))

    err := writer.WriteField("check_id", checkId)
    //_, err = io.Copy(part, file)
    err = writer.Close()

    header := c.GetBaseHeaders()
    header.Set("Content-Type", writer.FormDataContentType())
    jsonStr, _ := c.post(apiUploadUrl, body, header)
    var uploadData []UploadData
    err = unmarshal(jsonStr, &uploadData)

    return uploadData, err
}

func (c Codequiry) post(url string, data interface{}, header http.Header) (string, error) {
    var content string
    if data != nil && reflect.TypeOf(data).Kind() == reflect.Map {
        jsonStr, _  := json.Marshal(data)
        content = string(jsonStr)
    } else {
        if bytesBuff, ok := data.(*bytes.Buffer); ok {
            content = bytesBuff.String()
        }
    }
    var reqUrl string
    if strings.HasPrefix(url, "http") {
        reqUrl = url
    } else {
        reqUrl = apiBaseUrl + url
    }

    req, _ := http.NewRequest("POST", reqUrl, strings.NewReader(content))
    if header == nil {
        header = c.GetBaseHeaders()
    }
    req.Header = header

    client := &http.Client{}
    resp, err := client.Do(req)

    if resp != nil {
        bodyBytes, _ := ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }

        defer resp.Body.Close()
        return string(bodyBytes), nil
    } else {
        return "", nil
    }
}

func (sd *date) UnmarshalJSON(input []byte) error {
    strInput := string(input)
    strInput = strings.Trim(strInput, `"`)
    newTime, err := time.Parse("2006-01-02 15:04:05", strInput)
    if err != nil {
        return err
    }

    sd.Time = newTime
    return nil
}

func unmarshal(jsonStr string, target interface{}) error {
    _ = json.Unmarshal([]byte(jsonStr), target)
    if errorRegexp.MatchString(jsonStr) {
        msg := errorRegexp.FindStringSubmatch(jsonStr)
        var err = APIError{msg[1]}
        _ = json.Unmarshal([]byte(jsonStr), &err)
        return &err
    }

    return nil
}

func main() {
    var cq = Codequiry{ApiKey: ""}

    account, err := cq.Account()
    fmt.Printf("%+v\n%+v\n",account, err)
}

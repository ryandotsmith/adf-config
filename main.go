package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bmizerany/aws4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type DynamoClient struct {
	Url    string
	Client *aws4.Client
	Token  string
}

type ST struct {
	S string
}

type AVList struct {
	AttributeValueList []ST
	ComparisonOperator string
}

type AppCondition struct {
	App AVList
}

func (c DynamoClient) do(target string, in, out interface{}) error {
	j, err := json.Marshal(in)
	data := bytes.NewBuffer([]byte(j))
	r, _ := http.NewRequest("POST", c.Url, data)
	r.Header.Set("Content-Type", "application/x-amz-json-1.0")
	r.Header.Set("X-Amz-Target", target)
	r.Header.Set("x-amz-security-token", c.Token)
	resp, err := c.Client.Do(r)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return json.Unmarshal(respBody, out)
}

func list(dc *DynamoClient, app string) []string {
	req := struct {
		TableName      string
		ConsistentRead bool
		ScanFilter     AppCondition
	}{
		"adf-config",
		true,
		AppCondition{AVList{[]ST{ST{app}}, "EQ"}},
	}
	resp := &struct {
		Count int
		Items []struct {
			Name  ST
			Value ST
		}
	}{}
	if err := dc.do("DynamoDB_20120810.Scan", req, resp); err != nil {
		log.Fatal(err)
	}
	res := make([]string, resp.Count)
	for i, item := range resp.Items {
		res[i] = item.Name.S + "=" + item.Value.S
	}
	return res
}

// Using IAM Roles, we can retreive api credentials from
// the EC2 instances meta data. When running on EC2 there are
// no secrets required.
func loadLocalKeys() (*aws4.Keys, string) {
	h := "http://169.254.169.254"
	p := "/latest/meta-data/iam/security-credentials/adf-config"
	resp, err := http.Get(h + p)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	tmp := &struct {
		AccessKeyId, SecretAccessKey, Token string
	}{}
	json.Unmarshal(b, tmp)
	return &aws4.Keys{tmp.AccessKeyId, tmp.SecretAccessKey}, tmp.Token
}

// Fetching EC2's region
func loadRegion() string {
	h := "http://169.254.169.254"
	p := "/latest/meta-data/placement/availability-zone"
	resp, err := http.Get(h + p)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// Sometimes the output ends in a letter, which doesn't match the available regions:
	// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
	last := b[len(b)-1]
	if last != '1' || last != '2' {
		b = b[:len(b)-1]
	}
	return string(b)
}

func main() {
	var listCmd bool
	flag.BoolVar(&listCmd, "l", true, "list config")
	var app string
	flag.StringVar(&app, "a", "", "app name")
	flag.Parse()

	k, tok := loadLocalKeys()
	dc := new(DynamoClient)
	region := loadRegion()
	dc.Url = "https://dynamodb."+region+".amazonaws.com"
	dc.Client = &aws4.Client{Keys: k}
	dc.Token = tok

	if listCmd {
		for _, v := range list(dc, app) {
			fmt.Println(v)
		}
		os.Exit(0)
	}
}

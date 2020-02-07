package awslogs

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"encoding/csv"
	"strings"

	"go.uber.org/zap"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers/timestamp"
)

var S3ServerAccessDesc = `S3ServerAccess is an AWS S3 Access Log.
Log format & samples can be seen here: https://docs.aws.amazon.com/AmazonS3/latest/dev/LogFormat.html`

const (
	s3ServerAccessMinNumberOfColumns = 25
)

// nolint:lll
type S3ServerAccess struct {
	BucketOwner        *string            `json:"bucketowner,omitempty" validate:"required,len=64,alphanum" description:"The canonical user ID of the owner of the source bucket. The canonical user ID is another form of the AWS account ID."`
	Bucket             *string            `json:"bucket,omitempty" description:"The name of the bucket that the request was processed against. If the system receives a malformed request and cannot determine the bucket, the request will not appear in any server access log."`
	Time               *timestamp.RFC3339 `json:"time,omitempty" description:"The time at which the request was received (UTC)."`
	RemoteIP           *string            `json:"remoteip,omitempty" description:"The apparent internet address of the requester. Intermediate proxies and firewalls might obscure the actual address of the machine making the request."`
	Requester          *string            `json:"requester,omitempty" description:"The canonical user ID of the requester, or NULL for unauthenticated requests. If the requester was an IAM user, this field returns the requester's IAM user name along with the AWS root account that the IAM user belongs to. This identifier is the same one used for access control purposes."`
	RequestID          *string            `json:"requestid,omitempty" description:"A string generated by Amazon S3 to uniquely identify each request."`
	Operation          *string            `json:"operation,omitempty" description:"The operation listed here is declared as SOAP.operation, REST.HTTP_method.resource_type, WEBSITE.HTTP_method.resource_type, or BATCH.DELETE.OBJECT."`
	Key                *string            `json:"key,omitempty" description:"The key part of the request, URL encoded, or NULL if the operation does not take a key parameter."`
	RequestURI         *string            `json:"requesturi,omitempty" description:"The Request-URI part of the HTTP request message."`
	HTTPStatus         *int               `json:"httpstatus,omitempty" validate:"required,max=600,min=100" description:"The numeric HTTP status code of the response."`
	ErrorCode          *string            `json:"errorcode,omitempty" description:"The Amazon S3 Error Code, or NULL if no error occurred."`
	BytesSent          *int               `json:"bytessent,omitempty" description:"The number of response bytes sent, excluding HTTP protocol overhead, or NULL if zero."`
	ObjectSize         *int               `json:"objectsize,omitempty" description:"The total size of the object in question."`
	TotalTime          *int               `json:"totaltime,omitempty" description:"The number of milliseconds the request was in flight from the server's perspective. This value is measured from the time your request is received to the time that the last byte of the response is sent. Measurements made from the client's perspective might be longer due to network latency."`
	TurnAroundTime     *int               `json:"turnaroundtime,omitempty" description:"The number of milliseconds that Amazon S3 spent processing your request. This value is measured from the time the last byte of your request was received until the time the first byte of the response was sent."`
	Referrer           *string            `json:"referrer,omitempty" description:"The value of the HTTP Referer header, if present. HTTP user-agents (for example, browsers) typically set this header to the URL of the linking or embedding page when making a request."`
	UserAgent          *string            `json:"useragent,omitempty" description:"The value of the HTTP User-Agent header."`
	VersionID          *string            `json:"versionid,omitempty" description:"The version ID in the request, or NULL if the operation does not take a versionId parameter."`
	HostID             *string            `json:"hostid,omitempty" description:"The x-amz-id-2 or Amazon S3 extended request ID."`
	SignatureVersion   *string            `json:"signatureversion,omitempty" description:"The signature version, SigV2 or SigV4, that was used to authenticate the request or NULL for unauthenticated requests."`
	CipherSuite        *string            `json:"ciphersuite,omitempty" description:"The Secure Sockets Layer (SSL) cipher that was negotiated for HTTPS request or NULL for HTTP."`
	AuthenticationType *string            `json:"authenticationtype,omitempty" description:"The type of request authentication used, AuthHeader for authentication headers, QueryString for query string (pre-signed URL) or NULL for unauthenticated requests."`
	HostHeader         *string            `json:"hostheader,omitempty" description:"The endpoint used to connect to Amazon S3."`
	TLSVersion         *string            `json:"tlsVersion,omitempty" description:"The Transport Layer Security (TLS) version negotiated by the client. The value is one of following: TLSv1, TLSv1.1, TLSv1.2; or NULL if TLS wasn't used."`
	AdditionalFields   []string           `json:"additionalFields,omitempty" description:"The remaining columns in the record as an array."`

	// NOTE: added to end of struct to allow expansion later
	AWSPantherLog
}

// S3ServerAccessParser parses AWS S3 Server Access logs
type S3ServerAccessParser struct{}

func (p *S3ServerAccessParser) New() parsers.LogParser {
	return &S3ServerAccessParser{}
}

func (p *S3ServerAccessParser) ParseHeader(log string) []interface{} {
	return p.Parse(log) // no header
}

// Parse returns the parsed events or nil if parsing failed
func (p *S3ServerAccessParser) Parse(log string) []interface{} {
	reader := csv.NewReader(strings.NewReader(log))
	reader.LazyQuotes = true
	reader.Comma = ' '

	records, err := reader.ReadAll()
	if len(records) == 0 || err != nil {
		zap.L().Debug("failed to parse the log as csv")
		return nil
	}

	// parser should only receive 1 line at a time
	record := records[0]
	if len(record) < s3ServerAccessMinNumberOfColumns {
		zap.L().Debug("failed to parse the log as csv (wrong number of columns)")
		return nil
	}

	// The time in the logs is represented as [06/Feb/2019:00:00:38 +0000]
	// The CSV reader will break the above date to two different fields `[06/Feb/2019:00:00:38` and `+0000]`
	// We concatenate these fields before trying to parse them
	parsedTime, err := timestamp.Parse("[2/Jan/2006:15:04:05-0700]", record[2]+record[3])
	if err != nil {
		zap.L().Debug("failed to parse timestamp log as csv")
		return nil
	}

	var additionalFields []string = nil
	if len(record) > 25 {
		additionalFields = record[25:]
	}

	event := &S3ServerAccess{
		BucketOwner:        parsers.CsvStringToPointer(record[0]),
		Bucket:             parsers.CsvStringToPointer(record[1]),
		Time:               &parsedTime,
		RemoteIP:           parsers.CsvStringToPointer(record[4]),
		Requester:          parsers.CsvStringToPointer(record[5]),
		RequestID:          parsers.CsvStringToPointer(record[6]),
		Operation:          parsers.CsvStringToPointer(record[7]),
		Key:                parsers.CsvStringToPointer(record[8]),
		RequestURI:         parsers.CsvStringToPointer(record[9]),
		HTTPStatus:         parsers.CsvStringToIntPointer(record[10]),
		ErrorCode:          parsers.CsvStringToPointer(record[11]),
		BytesSent:          parsers.CsvStringToIntPointer(record[12]),
		ObjectSize:         parsers.CsvStringToIntPointer(record[13]),
		TotalTime:          parsers.CsvStringToIntPointer(record[14]),
		TurnAroundTime:     parsers.CsvStringToIntPointer(record[15]),
		Referrer:           parsers.CsvStringToPointer(record[16]),
		UserAgent:          parsers.CsvStringToPointer(record[17]),
		VersionID:          parsers.CsvStringToPointer(record[18]),
		HostID:             parsers.CsvStringToPointer(record[19]),
		SignatureVersion:   parsers.CsvStringToPointer(record[20]),
		CipherSuite:        parsers.CsvStringToPointer(record[21]),
		AuthenticationType: parsers.CsvStringToPointer(record[22]),
		HostHeader:         parsers.CsvStringToPointer(record[23]),
		TLSVersion:         parsers.CsvStringToPointer(record[24]),
		AdditionalFields:   additionalFields,
	}

	event.updatePantherFields(p)

	if err := parsers.Validator.Struct(event); err != nil {
		zap.L().Debug("failed to validate log", zap.Error(err))
		return nil
	}

	return []interface{}{event}
}

// LogType returns the log type supported by this parser
func (p *S3ServerAccessParser) LogType() string {
	return "AWS.S3ServerAccess"
}

func (event *S3ServerAccess) updatePantherFields(p *S3ServerAccessParser) {
	event.SetCoreFieldsPtr(p.LogType(), event.Time)
	event.AppendAnyIPAddressPtrs(event.RemoteIP)
	if event.Requester != nil && strings.HasPrefix(*event.Requester, "arn:") {
		event.AppendAnyAWSARNs(*event.Requester)
	}
}

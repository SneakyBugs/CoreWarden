package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sneakybugs/corewarden/client"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

type MockClientAction struct {
	action     any
	stateAfter []client.Record
}

type MockClient struct {
	actions            []MockClientAction
	currentActionIndex int
	t                  *testing.T
}

func (c *MockClient) CreateRecord(params client.CreateRecordParams) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called CreateRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(CreateRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method CreateRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.Zone != params.Zone {
		c.t.Fatalf("Expected params.Zone to be '%s', got '%s'\n", clientAction.Zone, params.Zone)
	}
	if clientAction.RR != params.RR {
		c.t.Fatalf("Expected params.RR to be '%s', got '%s'\n", clientAction.RR, params.RR)
	}
	if clientAction.Comment != params.Comment {
		c.t.Fatalf("Expected params.Comment to be '%s', got '%s'\n", clientAction.Comment, params.Comment)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type CreateRecordAction struct {
	Zone           string
	RR             string
	Comment        string
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) ReadRecord(id int) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called ReadRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(ReadRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method ReadRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != id {
		c.t.Fatalf("Expected id to be '%d', got '%d'\n", clientAction.ID, id)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type ReadRecordAction struct {
	ID             int
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) UpdateRecord(params client.UpdateRecordParams) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called UpdateRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(UpdateRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method UpdateRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != params.ID {
		c.t.Fatalf("Expected params.ID to be '%d', got '%d\n", clientAction.ID, params.ID)
	}
	if clientAction.Zone != params.Zone {
		c.t.Fatalf("Expected params.Zone to be '%s', got '%s'\n", clientAction.Zone, params.Zone)
	}
	if clientAction.RR != params.RR {
		c.t.Fatalf("Expected params.RR to be '%s', got '%s'\n", clientAction.RR, params.RR)
	}
	if clientAction.Comment != params.Comment {
		c.t.Fatalf("Expected params.Comment to be '%s', got '%s'\n", clientAction.Comment, params.Comment)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type UpdateRecordAction struct {
	ID             int
	Zone           string
	RR             string
	Comment        string
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) DeleteRecord(id int) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called DeleteRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(DeleteRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method DeleteRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != id {
		c.t.Fatalf("Expected id to be '%d', got '%d'\n", clientAction.ID, id)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type DeleteRecordAction struct {
	ID             int
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) ListRecords(zone string) ([]client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called ListRecords when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(ListRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method ListRecords during action %d\n", c.currentActionIndex)
	}
	if clientAction.Zone != zone {
		c.t.Fatalf("Expected zone to be '%s', got '%s'\n", clientAction.Zone, zone)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecords, clientAction.ResponseErr
}

type ListRecordAction struct {
	Zone            string
	ResponseRecords []client.Record
	ResponseErr     error
}

func newTestProvider(t *testing.T, actions []MockClientAction) Provider {
	mockClient := MockClient{
		actions:            actions,
		currentActionIndex: 0,
		t:                  t,
	}
	return Provider{
		client: &mockClient,
		zones:  []string{"example.com."},
	}
}

func (c *MockClient) IsFinished() bool {
	return c.currentActionIndex == len(c.actions)
}

func assertMockFinished(t *testing.T, c client.Client) {
	v, ok := c.(*MockClient)
	if !ok {
		t.Fatal("Expected c to be a MockClient, got other client.Client\n")
	}
	if !v.IsFinished() {
		t.Fatal("MockClient did not finish all actions\n")
	}
}

func createTestRecord(t *testing.T, id int, zone string, content string, comment string) client.Record {
	now := time.Now()
	rr, err := dns.NewRR(content)
	assert.Nil(t, err)
	return client.Record{
		ID:        id,
		Zone:      zone,
		RR:        rr,
		Comment:   comment,
		CreatedAt: now,
		UpdatedOn: now,
	}
}

func assertActions(t *testing.T, endpoints []*endpoint.Endpoint, actions []MockClientAction, managedRecords []string) {
	err := assertActionsE(t, endpoints, actions, managedRecords)
	assert.NoError(t, err)
}

func assertActionsE(t *testing.T, endpoints []*endpoint.Endpoint, actions []MockClientAction, managedRecords []string) error {
	p := newTestProvider(t, actions)
	records, err := p.Records(context.TODO())
	assert.NoError(t, err)
	endpoints, err = p.AdjustEndpoints(endpoints)
	assert.NoError(t, err)
	domainFilter := endpoint.NewDomainFilter([]string{"example.com"})
	plan := &plan.Plan{
		Current:        records,
		Desired:        endpoints,
		DomainFilter:   endpoint.MatchAllDomainFilters{&domainFilter},
		ManagedRecords: managedRecords,
	}
	changes := plan.Calculate().Changes
	err = p.ApplyChanges(context.TODO(), changes)
	assertMockFinished(t, p.client)
	return err
}

func TestNewARecords(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone:            "example.com.",
				ResponseRecords: []client.Record{},
				ResponseErr:     nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
				ResponseErr:    nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestNewARecordsSubdomain(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "foo.bar.example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone:            "example.com.",
				ResponseRecords: []client.Record{},
				ResponseErr:     nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "foo.bar.\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 1, "foo.bar.example.com.", "foo.bar. 0 IN A 10.0.0.1", ""),
				ResponseErr:    nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "foo.bar.\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "foo.bar.example.com.", "foo.bar. 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestNewARecordsSingleSubdomain(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "my-app.example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone:            "example.com.",
				ResponseRecords: []client.Record{},
				ResponseErr:     nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "my-app.\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 1, "my-app.example.com.", "my-app. 0 IN A 10.0.0.1", ""),
				ResponseErr:    nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "my-app.\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "my-app.example.com.", "my-app. 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestNewTargetInExistingARecord(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
					createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
					createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.3",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 3, "example.com.", ". 0 IN A 10.0.0.3", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestRemoveTargetInExistingARecord(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "example.com",
			Targets:    endpoint.Targets{"10.0.0.1"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
					createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
					createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: DeleteRecordAction{
				ID:             2,
				ResponseRecord: createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestUpdateTargetInExistingARecord(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 3600 IN A 10.0.0.1", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 3600 IN A 10.0.0.1", ""),
				},
				ResponseErr: nil,
			},
		},
		// Must also add or remove target for planner mark endpoint as changed.
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "example.com.", ". 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
		{
			action: UpdateRecordAction{
				ID:             1,
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
				ResponseErr:    nil,
			},
		},
	}
	assertActions(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
}

func TestNewEndpointRecoverableError(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "foo.bar.example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone:            "example.com.",
				ResponseRecords: []client.Record{},
				ResponseErr:     nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "foo.bar.\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: client.Record{},
				ResponseErr:    fmt.Errorf("Some error"),
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             "foo.bar.\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "foo.bar.example.com.", "foo.bar. 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	err := assertActionsE(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
	assert.ErrorContains(t, err, "Encountered 1 recoverable errors")
}

func TestUpdatedEndpointRecoverableError(t *testing.T) {
	endpoints := []*endpoint.Endpoint{
		{
			RecordType: "A",
			DNSName:    "example.com",
			Targets:    endpoint.Targets{"10.0.0.1", "10.0.0.2", "10.0.0.3"},
		},
	}
	actions := []MockClientAction{
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 In A 10.0.0.3", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: ListRecordAction{
				Zone: "example.com.",
				ResponseRecords: []client.Record{
					createTestRecord(t, 1, "example.com.", ". 0 In A 10.0.0.3", ""),
				},
				ResponseErr: nil,
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.1",
				Comment:        "",
				ResponseRecord: client.Record{},
				ResponseErr:    fmt.Errorf("Some error"),
			},
		},
		{
			action: CreateRecordAction{
				Zone:           "example.com.",
				RR:             ".\t0\tIN\tA\t10.0.0.2",
				Comment:        "",
				ResponseRecord: createTestRecord(t, 2, "foo.bar.example.com.", "foo.bar. 0 IN A 10.0.0.2", ""),
				ResponseErr:    nil,
			},
		},
	}
	err := assertActionsE(t, endpoints, actions, []string{endpoint.RecordTypeA, endpoint.RecordTypeCNAME})
	assert.ErrorContains(t, err, "Encountered 1 recoverable errors")
}

func TestRecordsBasic(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
		createTestRecord(t, 2, "example.com.", "@ 0 IN A 10.0.0.2", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com.",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints), "endpoints length should be 1")
	assert.Equal(t, 2, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "10.0.0.1", endpoints[0].Targets[0])
	assert.Equal(t, "10.0.0.2", endpoints[0].Targets[1])
	assert.Equal(t, "example.com", endpoints[0].DNSName)
}

func TestRecordsSubdomain(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", "foo 0 IN A 10.0.0.1", ""),
		createTestRecord(t, 2, "example.com.", "foo 0 IN A 10.0.0.2", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com.",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints), "endpoints length should be 1")
	assert.Equal(t, 2, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "10.0.0.1", endpoints[0].Targets[0])
	assert.Equal(t, "10.0.0.2", endpoints[0].Targets[1])
	assert.Equal(t, "foo.example.com", endpoints[0].DNSName)
}

func TestRecordsIgnoresUnsupportedTypes(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", ". 0 IN A 10.0.0.1", ""),
		createTestRecord(t, 2, "example.com.", ". 0 IN MX 10 mail.example.com.", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com.",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints), "endpoints length should be 2")
	assert.Equal(t, 1, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "10.0.0.1", endpoints[0].Targets[0])
	assert.Equal(t, "example.com", endpoints[0].DNSName)
}

func TestRecordsTXTRecordTarget(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", ". 0 IN TXT \"Hello\nworld\tsome\rmore space\" no space between", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com.",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints), "endpoints length should be 2")
	assert.Equal(t, 1, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "Hello\nworld\tsome\rmore spacenospacebetween", endpoints[0].Targets[0])
}

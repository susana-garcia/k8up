package observe

import (
	"reflect"
	"sync"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
)

func TestBroker_Subscribe(t *testing.T) {
	type fields struct {
		subscribers map[topic][]Subscriber
	}
	type args struct {
		topicName string
	}
	tests := []struct {
		name    string
		loops   int
		fields  fields
		args    args
		want    *Subscriber
		wantErr bool
	}{
		{
			name:  "Subscribe without duplicate on existing",
			loops: 10000,
			fields: fields{
				map[topic][]Subscriber{
					"testTopic": {
						{
							id:    5577006791947779410,
							mutex: &sync.Mutex{},
						},
					},
				},
			},
			args: args{
				topicName: "testTopic",
			},
			want: &Subscriber{
				CH:    make(chan PodState),
				mutex: &sync.Mutex{},
			},
		},
		{
			name:  "Subscribe without duplicate on new",
			loops: 10000,
			fields: fields{
				subscribers: make(map[topic][]Subscriber, 0),
			},
			args: args{
				topicName: "testTopic",
			},
			want: &Subscriber{
				CH: make(chan PodState),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Broker{
				subscribers: tt.fields.subscribers,
				mutex:       &sync.Mutex{},
			}
			for i := 0; i < tt.loops; i++ {
				got, err := b.Subscribe(tt.args.topicName)
				if (err != nil) != tt.wantErr {
					t.Errorf("Broker.Subscribe() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if reflect.DeepEqual(got.id, tt.want.id) {
					t.Errorf("Broker.Subscribe() = %v", got.id)
				}
			}
		})
	}
}

func TestBroker_Unsubscribe(t *testing.T) {
	type fields struct {
		subscribers map[topic][]Subscriber
	}
	type args struct {
		topicName  string
		subscriber *Subscriber
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[topic][]Subscriber
	}{
		{
			name: "Unsubscribe",
			want: map[topic][]Subscriber{
				"test": {},
			},
			fields: fields{
				subscribers: map[topic][]Subscriber{
					"test": {
						{
							CH:    make(chan PodState, 0),
							id:    512,
							mutex: &sync.Mutex{},
						},
					},
				},
			},
			args: args{
				topicName: "test",
				subscriber: &Subscriber{
					CH:    make(chan PodState, 0),
					id:    512,
					mutex: &sync.Mutex{},
				},
			},
		},
		{
			name: "Unsubscribe on empty broker",
			want: map[topic][]Subscriber{},
			fields: fields{
				subscribers: map[topic][]Subscriber{},
			},
			args: args{
				topicName: "test",
				subscriber: &Subscriber{
					CH:    make(chan PodState, 0),
					id:    512,
					mutex: &sync.Mutex{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Broker{
				subscribers: tt.fields.subscribers,
				mutex:       &sync.Mutex{},
			}
			b.Unsubscribe(tt.args.topicName, tt.args.subscriber)

			if !reflect.DeepEqual(tt.fields.subscribers, tt.want) {
				t.Errorf("Broker.Subscribe() = %v, want %v", tt.fields.subscribers, tt.want)
			}

		})
	}
}

func TestBroker_Notify(t *testing.T) {
	type fields struct {
		subscribers map[topic][]Subscriber
	}
	type args struct {
		topicName string
		state     PodState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test notify on existing topic",
			fields: fields{
				subscribers: map[topic][]Subscriber{
					"test": []Subscriber{
						{
							CH:    make(chan PodState, 0),
							id:    1337,
							mutex: &sync.Mutex{},
						},
					},
				},
			},
			args: args{
				topicName: "test",
				state: PodState{
					BaasID:     "test",
					Repository: "test",
					State:      batchv1.JobFailed,
				},
			},
		},
		{
			name:    "Test notify on non-existing topic",
			wantErr: true,
			fields: fields{
				subscribers: make(map[topic][]Subscriber, 0),
			},
			args: args{
				topicName: "test",
				state: PodState{
					BaasID:     "test",
					Repository: "test",
					State:      batchv1.JobFailed,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Broker{
				subscribers: tt.fields.subscribers,
				mutex:       &sync.Mutex{},
			}
			if err := b.Notify(tt.args.topicName, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("Broker.Notify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

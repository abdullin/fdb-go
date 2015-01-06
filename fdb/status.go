// FoundationDB Go API
// Copyright (c) 2013 FoundationDB, LLC

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package fdb

// Status information.
// Generated with https://github.com/bemasher/JSONGen
type Status struct {
	Client struct {
		ClusterFile struct {
			Path     string `json:"path"`
			UpToDate bool   `json:"up_to_date"`
		} `json:"cluster_file"`
		Coordinators struct {
			Coordinators []struct {
				Address   string `json:"address"`
				Reachable bool   `json:"reachable"`
			} `json:"coordinators"`
			QuorumReachable bool `json:"quorum_reachable"`
		} `json:"coordinators"`
		DatabaseStatus struct {
			Available bool `json:"available"`
			Healthy   bool `json:"healthy"`
		} `json:"database_status"`
		Messages []struct {
			Description string `json:"description"`
			Name        string `json:"name"`
		} `json:"messages"`
		Timestamp int64 `json:"timestamp"`
	} `json:"client"`
	Cluster struct {
		ClusterControllerTimestamp int64 `json:"cluster_controller_timestamp"`
		Configuration              struct {
			CoordinatorsCount int64 `json:"coordinators_count"`
			ExcludedServers   []struct {
				Address string `json:"address"`
			} `json:"excluded_servers"`
			Logs       int64 `json:"logs"`
			Proxies    int64 `json:"proxies"`
			Redundancy struct {
				Factor string `json:"factor"`
			} `json:"redundancy"`
			Resolvers     int64  `json:"resolvers"`
			StorageEngine string `json:"storage_engine"`
		} `json:"configuration"`
		Data struct {
			AveragePartitionSizeBytes             int64 `json:"average_partition_size_bytes"`
			LeastOperatingSpaceBytesLogServer     int64 `json:"least_operating_space_bytes_log_server"`
			LeastOperatingSpaceBytesStorageServer int64 `json:"least_operating_space_bytes_storage_server"`
			MovingData                            struct {
				InFlightBytes int64 `json:"in_flight_bytes"`
				InQueueBytes  int64 `json:"in_queue_bytes"`
			} `json:"moving_data"`
			PartitionsCount int64 `json:"partitions_count"`
			State           struct {
				Description          string `json:"description"`
				Healthy              bool   `json:"healthy"`
				MinReplicasRemaining int64  `json:"min_replicas_remaining"`
				Name                 string `json:"name"`
			} `json:"state"`
			TotalDiskUsedBytes int64 `json:"total_disk_used_bytes"`
			TotalKvSizeBytes   int64 `json:"total_kv_size_bytes"`
		} `json:"data"`
		FaultTolerance struct {
			MaxMachineFailuresWithoutLosingAvailability int64 `json:"max_machine_failures_without_losing_availability"`
			MaxMachineFailuresWithoutLosingData         int64 `json:"max_machine_failures_without_losing_data"`
		} `json:"fault_tolerance"`
		LatencyProbe struct {
			CommitSeconds           float64 `json:"commit_seconds"`
			ReadSeconds             float64 `json:"read_seconds"`
			TransactionStartSeconds float64 `json:"transaction_start_seconds"`
		} `json:"latency_probe"`
		License  string `json:"license"`
		Machines map[string]struct {
			Address string `json:"address"`
			Cpu     struct {
				LogicalCoreUtilization float64 `json:"logical_core_utilization"`
			} `json:"cpu"`
			DatacenterId string `json:"datacenter_id"`
			Excluded     bool   `json:"excluded"`
			MachineId    string `json:"machine_id"`
			Memory       struct {
				CommittedBytes int64 `json:"committed_bytes"`
				FreeBytes      int64 `json:"free_bytes"`
				TotalBytes     int64 `json:"total_bytes"`
			} `json:"memory"`
			Network struct {
				MegabitsReceived struct {
					Hz float64 `json:"hz"`
				} `json:"megabits_received"`
				MegabitsSent struct {
					Hz float64 `json:"hz"`
				} `json:"megabits_sent"`
				TcpSegmentsRetransmitted struct {
					Hz float64 `json:"hz"`
				} `json:"tcp_segments_retransmitted"`
			} `json:"network"`
		} `json:"machines"`
		Messages []struct {
			Description string `json:"description"`
			Issues      []struct {
				Description string `json:"description"`
				Name        string `json:"name"`
			} `json:"issues"`
			Name    string `json:"name"`
			Reasons []struct {
				Description string `json:"description"`
			} `json:"reasons"`
			UnreachableProcesses []struct {
				Address string `json:"address"`
			} `json:"unreachable_processes"`
		} `json:"messages"`
		Processes map[string]struct {
			Address     string `json:"address"`
			CommandLine string `json:"command_line"`
			Cpu         struct {
				UsageCores float64 `json:"usage_cores"`
			} `json:"cpu"`
			Disk struct {
				Busy float64 `json:"busy"`
			} `json:"disk"`
			Excluded  bool   `json:"excluded"`
			MachineId string `json:"machine_id"`
			Memory    struct {
				AvailableBytes int64 `json:"available_bytes"`
				UsedBytes      int64 `json:"used_bytes"`
			} `json:"memory"`
			Messages []struct {
				Description   string  `json:"description"`
				Name          string  `json:"name"`
				RawLogMessage string  `json:"raw_log_message"`
				Time          float64 `json:"time"`
				Type          string  `json:"type"`
			} `json:"messages"`
			Network struct {
				MegabitsReceived struct {
					Hz float64 `json:"hz"`
				} `json:"megabits_received"`
				MegabitsSent struct {
					Hz float64 `json:"hz"`
				} `json:"megabits_sent"`
			} `json:"network"`
			Roles []struct {
				Id   string `json:"id"`
				Role string `json:"role"`
			} `json:"roles"`
			Version string `json:"version"`
		} `json:"processes"`
		Qos struct {
			PerformanceLimitedBy struct {
				Description    string `json:"description"`
				Name           string `json:"name"`
				ReasonServerId string `json:"reason_server_id"`
			} `json:"performance_limited_by"`
			WorstQueueBytesLogServer     int64 `json:"worst_queue_bytes_log_server"`
			WorstQueueBytesStorageServer int64 `json:"worst_queue_bytes_storage_server"`
		} `json:"qos"`
		RecoveryState struct {
			Description       string `json:"description"`
			Name              string `json:"name"`
			RequiredLogs      int64  `json:"required_logs"`
			RequiredProxies   int64  `json:"required_proxies"`
			RequiredResolvers int64  `json:"required_resolvers"`
		} `json:"recovery_state"`
		Workload struct {
			Bytes struct {
				Written struct {
					Counter   int64   `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"written"`
			} `json:"bytes"`
			Operations struct {
				Reads struct {
					Hz float64 `json:"hz"`
				} `json:"reads"`
				Writes struct {
					Counter   int64   `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"writes"`
			} `json:"operations"`
			Transactions struct {
				Committed struct {
					Counter   int64   `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"committed"`
				Conflicted struct {
					Counter   int64   `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"conflicted"`
				Started struct {
					Counter   int64   `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"started"`
			} `json:"transactions"`
		} `json:"workload"`
	} `json:"cluster"`
}

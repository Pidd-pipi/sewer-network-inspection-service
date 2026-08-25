package main

import (
	"sort"
	"strings"
)

func formatRecordLine(item OpsRecord) string {
	return strings.Join([]string{
		item.ID,
		item.Subject,
		item.Owner,
		string(item.Status),
		string(item.Priority),
		item.UpdatedAt,
	}, ",")
}

func exportRecordLines(items []OpsRecord) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		lines = append(lines, formatRecordLine(item))
	}
	return lines
}

func exportRecordsCSV(items []OpsRecord) string {
	if len(items) == 0 {
		return ""
	}
	ordered := make([]OpsRecord, len(items))
	copy(ordered, items)
	sortOpsRecords(ordered)
	return strings.Join(exportRecordLines(ordered), "\n")
}

func exportSnapshotCSV(snapshot OpsSnapshot) string {
	var builder strings.Builder
	builder.WriteString("domain,generated_at,records,active\n")
	builder.WriteString(strings.Join([]string{
		snapshot.Domain,
		snapshot.GeneratedAt,
		formatOpsInt(snapshot.Records),
		formatOpsInt(snapshot.Active),
	}, ","))
	builder.WriteString("\n")
	statuses := make([]OpsStatus, 0, len(snapshot.ByStatus))
	for status := range snapshot.ByStatus {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i] < statuses[j] })
	for _, status := range statuses {
		builder.WriteString(strings.Join([]string{"status", string(status), formatOpsInt(snapshot.ByStatus[status])}, ","))
		builder.WriteString("\n")
	}
	return builder.String()
}

func exportIDsCSV(items []OpsRecord) string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return strings.Join(ids, ",")
}

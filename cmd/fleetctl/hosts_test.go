package main

import (
	"context"
	"testing"

	"github.com/fleetdm/fleet/v4/server/fleet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostTransferFlagChecks(t *testing.T) {
	runServerWithMockedDS(t)

	runAppCheckErr(t,
		[]string{"hosts", "transfer", "--team", "team1", "--hosts", "host1", "--label", "AAA"},
		"--hosts cannot be used along side any other flag",
	)
	runAppCheckErr(t,
		[]string{"hosts", "transfer", "--team", "team1"},
		"You need to define either --hosts, or one or more of --label, --status, --search_query",
	)
}

func TestHostsTransferByHosts(t *testing.T) {
	_, ds := runServerWithMockedDS(t)

	ds.HostByIdentifierFunc = func(ctx context.Context, identifier string) (*fleet.Host, error) {
		require.Equal(t, "host1", identifier)
		return &fleet.Host{ID: 42}, nil
	}

	ds.TeamByNameFunc = func(ctx context.Context, name string) (*fleet.Team, error) {
		require.Equal(t, "team1", name)
		return &fleet.Team{ID: 99, Name: "team1"}, nil
	}

	ds.AddHostsToTeamFunc = func(ctx context.Context, teamID *uint, hostIDs []uint) error {
		require.NotNil(t, teamID)
		require.Equal(t, uint(99), *teamID)
		require.Equal(t, []uint{42}, hostIDs)
		return nil
	}

	ds.ReconcileProfilesOnTeamChangeFunc = func(ctx context.Context, hostIDs []uint, teamID *uint) error {
		return nil
	}

	ds.BulkSetPendingMDMAppleHostProfilesFunc = func(ctx context.Context, hostIDs, teamIDs, profileIDs []uint, uuids []string) error {
		return nil
	}

	assert.Equal(t, "", runAppForTest(t, []string{"hosts", "transfer", "--team", "team1", "--hosts", "host1"}))
}

func TestHostsTransferByLabel(t *testing.T) {
	_, ds := runServerWithMockedDS(t)

	ds.HostByIdentifierFunc = func(ctx context.Context, identifier string) (*fleet.Host, error) {
		require.Equal(t, "host1", identifier)
		return &fleet.Host{ID: 42}, nil
	}

	ds.TeamByNameFunc = func(ctx context.Context, name string) (*fleet.Team, error) {
		require.Equal(t, "team1", name)
		return &fleet.Team{ID: 99, Name: "team1"}, nil
	}

	ds.LabelIDsByNameFunc = func(ctx context.Context, labels []string) ([]uint, error) {
		require.Equal(t, []string{"label1"}, labels)
		return []uint{uint(11)}, nil
	}

	ds.ListHostsInLabelFunc = func(ctx context.Context, filter fleet.TeamFilter, lid uint, opt fleet.HostListOptions) ([]*fleet.Host, error) {
		require.Equal(t, fleet.HostStatus(""), opt.StatusFilter)
		require.Equal(t, uint(11), lid)
		return []*fleet.Host{{ID: 32}, {ID: 12}}, nil
	}

	ds.AddHostsToTeamFunc = func(ctx context.Context, teamID *uint, hostIDs []uint) error {
		require.NotNil(t, teamID)
		require.Equal(t, uint(99), *teamID)
		require.Equal(t, []uint{32, 12}, hostIDs)
		return nil
	}

	ds.ReconcileProfilesOnTeamChangeFunc = func(ctx context.Context, hostIDs []uint, teamID *uint) error {
		return nil
	}

	ds.BulkSetPendingMDMAppleHostProfilesFunc = func(ctx context.Context, hostIDs, teamIDs, profileIDs []uint, uuids []string) error {
		return nil
	}

	assert.Equal(t, "", runAppForTest(t, []string{"hosts", "transfer", "--team", "team1", "--label", "label1"}))
}

func TestHostsTransferByStatus(t *testing.T) {
	_, ds := runServerWithMockedDS(t)

	ds.HostByIdentifierFunc = func(ctx context.Context, identifier string) (*fleet.Host, error) {
		require.Equal(t, "host1", identifier)
		return &fleet.Host{ID: 42}, nil
	}

	ds.TeamByNameFunc = func(ctx context.Context, name string) (*fleet.Team, error) {
		require.Equal(t, "team1", name)
		return &fleet.Team{ID: 99, Name: "team1"}, nil
	}

	ds.LabelIDsByNameFunc = func(ctx context.Context, labels []string) ([]uint, error) {
		require.Equal(t, []string{"label1"}, labels)
		return []uint{uint(11)}, nil
	}

	ds.ListHostsFunc = func(ctx context.Context, filter fleet.TeamFilter, opt fleet.HostListOptions) ([]*fleet.Host, error) {
		require.Equal(t, fleet.StatusOnline, opt.StatusFilter)
		return []*fleet.Host{{ID: 32}, {ID: 12}}, nil
	}

	ds.AddHostsToTeamFunc = func(ctx context.Context, teamID *uint, hostIDs []uint) error {
		require.NotNil(t, teamID)
		require.Equal(t, uint(99), *teamID)
		require.Equal(t, []uint{32, 12}, hostIDs)
		return nil
	}

	ds.ReconcileProfilesOnTeamChangeFunc = func(ctx context.Context, hostIDs []uint, teamID *uint) error {
		return nil
	}

	ds.BulkSetPendingMDMAppleHostProfilesFunc = func(ctx context.Context, hostIDs, teamIDs, profileIDs []uint, uuids []string) error {
		return nil
	}

	assert.Equal(t, "", runAppForTest(t,
		[]string{"hosts", "transfer", "--team", "team1", "--status", "online"}))
}

func TestHostsTransferByStatusAndSearchQuery(t *testing.T) {
	_, ds := runServerWithMockedDS(t)

	ds.HostByIdentifierFunc = func(ctx context.Context, identifier string) (*fleet.Host, error) {
		require.Equal(t, "host1", identifier)
		return &fleet.Host{ID: 42}, nil
	}

	ds.TeamByNameFunc = func(ctx context.Context, name string) (*fleet.Team, error) {
		require.Equal(t, "team1", name)
		return &fleet.Team{ID: 99, Name: "team1"}, nil
	}

	ds.LabelIDsByNameFunc = func(ctx context.Context, labels []string) ([]uint, error) {
		require.Equal(t, []string{"label1"}, labels)
		return []uint{uint(11)}, nil
	}

	ds.ListHostsFunc = func(ctx context.Context, filter fleet.TeamFilter, opt fleet.HostListOptions) ([]*fleet.Host, error) {
		require.Equal(t, fleet.StatusOnline, opt.StatusFilter)
		require.Equal(t, "somequery", opt.MatchQuery)
		return []*fleet.Host{{ID: 32}, {ID: 12}}, nil
	}

	ds.AddHostsToTeamFunc = func(ctx context.Context, teamID *uint, hostIDs []uint) error {
		require.NotNil(t, teamID)
		require.Equal(t, uint(99), *teamID)
		require.Equal(t, []uint{32, 12}, hostIDs)
		return nil
	}

	ds.ReconcileProfilesOnTeamChangeFunc = func(ctx context.Context, hostIDs []uint, teamID *uint) error {
		return nil
	}

	ds.BulkSetPendingMDMAppleHostProfilesFunc = func(ctx context.Context, hostIDs, teamIDs, profileIDs []uint, uuids []string) error {
		return nil
	}

	assert.Equal(t, "", runAppForTest(t,
		[]string{"hosts", "transfer", "--team", "team1", "--status", "online", "--search_query", "somequery"}))
}

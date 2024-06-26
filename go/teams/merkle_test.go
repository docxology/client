package teams

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/keybase/client/go/kbtest"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/keybase1"
	"github.com/keybase/client/go/teams/hidden"
	"github.com/stretchr/testify/require"
)

// Test getting the merkle leaf from the server.
// This is a test of MerkleClient.
func TestMerkle(t *testing.T) {
	tc := SetupTest(t, "team", 1)
	defer tc.Cleanup()

	_, err := kbtest.CreateAndSignupFakeUser("team", tc.G)
	require.NoError(t, err)

	name := createTeam(tc)

	team, err := GetForTestByStringName(context.TODO(), tc.G, name)
	require.NoError(t, err)

	leaf, err := tc.G.MerkleClient.LookupTeam(libkb.NewMetaContextForTest(tc), team.ID)
	require.NoError(t, err)
	require.NotNil(t, leaf)
	t.Logf("team merkle leaf: %v", spew.Sdump(leaf))
	if leaf.TeamID.IsNil() {
		t.Fatalf("nil teamID; likely merkle hasn't yet published and polling is busted")
	}
	require.Equal(t, team.ID, leaf.TeamID, "team id")
	require.Equal(t, team.chain().GetLatestSeqno(), leaf.Private.Seqno)
	require.Equal(t, team.chain().GetLatestLinkID(), leaf.Private.LinkID.Export())
	// leaf.Private.SigID not checked
	require.Nil(t, leaf.Public, "team public leaf")
}

func TestMerkleWithHidden(t *testing.T) {
	tc := SetupTest(t, "team", 1)
	defer tc.Cleanup()

	_, err := kbtest.CreateAndSignupFakeUser("team", tc.G)
	require.NoError(t, err)

	name := createTeam(tc)

	team, err := GetForTestByStringName(context.TODO(), tc.G, name)
	require.NoError(t, err)

	leaf, hiddenResp, lastMerkleRoot, err := tc.G.MerkleClient.LookupTeamWithHidden(libkb.NewMetaContextForTest(tc), team.ID, hidden.ProcessHiddenResponseFunc)
	require.NoError(t, err)
	require.NotNil(t, lastMerkleRoot)
	require.NotNil(t, leaf)
	t.Logf("team merkle leaf: %v", spew.Sdump(leaf))
	if leaf.TeamID.IsNil() {
		t.Fatalf("nil teamID; likely merkle hasn't yet published and polling is busted")
	}
	require.Equal(t, team.ID, leaf.TeamID, "team id")
	require.Equal(t, team.chain().GetLatestSeqno(), leaf.Private.Seqno)
	require.Equal(t, team.chain().GetLatestLinkID(), leaf.Private.LinkID.Export())
	// leaf.Private.SigID not checked
	require.Nil(t, leaf.Public, "team public leaf")
	require.Equal(t, libkb.MerkleHiddenResponseTypeOK, hiddenResp.RespType)
	require.EqualValues(t, 0, hiddenResp.UncommittedSeqno)

	// make a hidden rotation to later check that merkle/path returns the appropriate result
	err = team.Rotate(context.TODO(), keybase1.RotationType_HIDDEN)
	require.NoError(t, err)
	// reload the team after the hidden rotation
	team, err = GetForTestByStringName(context.TODO(), tc.G, name)
	require.NoError(t, err)

	leaf, hiddenResp, lastMerkleRoot, err = tc.G.MerkleClient.LookupTeamWithHidden(libkb.NewMetaContextForTest(tc), team.ID, hidden.ProcessHiddenResponseFunc)
	require.NoError(t, err)
	require.NotNil(t, lastMerkleRoot)
	require.NotNil(t, leaf)
	require.Equal(t, team.ID, leaf.TeamID, "team id mismatch")
	require.Equal(t, team.chain().GetLatestSeqno(), leaf.Private.Seqno)
	require.Equal(t, team.chain().GetLatestLinkID(), leaf.Private.LinkID.Export())
	require.True(t, hiddenResp.RespType == libkb.MerkleHiddenResponseTypeOK)
	require.EqualValues(t, 1, hiddenResp.UncommittedSeqno)

	requestNewBlindTreeFromArchitectAndWaitUntilDone(t, &tc)

	leaf, hiddenResp, lastMerkleRoot, err = tc.G.MerkleClient.LookupTeamWithHidden(libkb.NewMetaContextForTest(tc), team.ID, hidden.ProcessHiddenResponseFunc)
	require.NoError(t, err)
	require.NotNil(t, lastMerkleRoot)
	require.NotNil(t, leaf)
	require.Equal(t, team.ID, leaf.TeamID, "team id")
	require.Equal(t, team.chain().GetLatestSeqno(), leaf.Private.Seqno)
	require.Equal(t, team.chain().GetLatestLinkID(), leaf.Private.LinkID.Export())
	require.True(t, hiddenResp.RespType == libkb.MerkleHiddenResponseTypeOK)
	require.EqualValues(t, 1, hiddenResp.UncommittedSeqno)
	require.NotNil(t, team.HiddenChain(), "NIL hidden chain")
}

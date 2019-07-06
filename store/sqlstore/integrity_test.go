// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
)

func createChannel(ss store.Store, TeamId, CreatorId string) *model.Channel {
	m := model.Channel{}
	m.TeamId = TeamId
	m.CreatorId = CreatorId
	m.DisplayName = "Name"
	m.Name = "zz" + model.NewId() + "b"
	m.Type = model.CHANNEL_OPEN
	c, _ := ss.Channel().Save(&m, -1)
	return c
}

func createChannelWithTeamId(ss store.Store, id string) *model.Channel {
	return createChannel(ss, id, model.NewId());
}

func createChannelWithCreatorId(ss store.Store, id string) *model.Channel {
	return createChannel(ss, model.NewId(), id);
}

func createPost(ss store.Store, ChannelId, UserId string) *model.Post {
	m := model.Post{}
	m.ChannelId = ChannelId
	m.UserId = UserId
	m.Message = "zz" + model.NewId() + "b"
	p, _ := ss.Post().Save(&m)
	return p
}

func createPostWithChannelId(ss store.Store, id string) *model.Post {
	return createPost(ss, id, model.NewId());
}

func createPostWithUserId(ss store.Store, id string) *model.Post {
	return createPost(ss, model.NewId(), id);
}

func TestCheckIntegrity(t *testing.T) {
	StoreTest(t, func(t *testing.T, ss store.Store) {
		ss.DropAllTables()
		t.Run("there should be no orphaned records on new db", func(t *testing.T) {
			results := ss.CheckIntegrity()
			require.NotNil(t, results)
			for result := range results {
				require.IsType(t, store.IntegrityCheckResult{}, result)
				require.Nil(t, result.Err)
				require.Empty(t, result.Records)
			}
		})
	})
}

func TestCheckChannelsIntegrity(t *testing.T) {
	StoreTest(t, func(t *testing.T, ss store.Store) {
		sqlStore := ss.(*store.LayeredStore).DatabaseLayer.(SqlStore)
		dbmap := sqlStore.GetMaster()

		ss.DropAllTables()
		dbmap.DropTable(model.Channel{})

		t.Run("should fail with an error", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkChannelsIntegrity(dbmap, results)
			result := <-results
			require.NotNil(t, result.Err)
			close(results)
		})

		dbmap.CreateTablesIfNotExists()

		t.Run("should generate a report with no records", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkUsersIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Empty(t, result.Records)
			close(results)
		})

		t.Run("should generate a report with one record", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			post := createPostWithChannelId(ss, model.NewId())
			go checkChannelsIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Len(t, result.Records, 1)
			require.Equal(t, store.OrphanedRecord{
				ParentId: post.ChannelId,
				ChildId: post.Id,
			}, result.Records[0])
			close(results)
		})
	})
}

func TestCheckUsersIntegrity(t *testing.T) {
	StoreTest(t, func(t *testing.T, ss store.Store) {
		sqlStore := ss.(*store.LayeredStore).DatabaseLayer.(SqlStore)
		dbmap := sqlStore.GetMaster()

		ss.DropAllTables()
		dbmap.DropTable(model.User{})

		t.Run("should fail with an error", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkUsersIntegrity(dbmap, results)
			result := <-results
			require.NotNil(t, result.Err)
			close(results)
		})

		dbmap.CreateTablesIfNotExists()

		t.Run("should generate a report with no records", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkUsersIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Empty(t, result.Records)
			close(results)
		})

		t.Run("should generate a report with one record", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			post := createPostWithUserId(ss, model.NewId())
			go checkUsersIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Len(t, result.Records, 1)
			require.Equal(t, store.OrphanedRecord{
				ParentId: post.UserId,
				ChildId: post.Id,
			}, result.Records[0])
			close(results)
		})
	})
}

func TestCheckTeamsIntegrity(t *testing.T) {
	StoreTest(t, func(t *testing.T, ss store.Store) {
		sqlStore := ss.(*store.LayeredStore).DatabaseLayer.(SqlStore)
		dbmap := sqlStore.GetMaster()

		ss.DropAllTables()
		dbmap.DropTable(model.Team{})

		t.Run("should fail with an error", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkTeamsIntegrity(dbmap, results)
			result := <-results
			require.NotNil(t, result.Err)
			close(results)
		})

		dbmap.CreateTablesIfNotExists()

		t.Run("should generate a report with no records", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			go checkTeamsIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Empty(t, result.Records)
			close(results)
		})

		t.Run("should generate a report with one record", func(t *testing.T) {
			results := make(chan store.IntegrityCheckResult)
			channel := createChannelWithTeamId(ss, model.NewId())
			go checkTeamsIntegrity(dbmap, results)
			result := <-results
			require.Nil(t, result.Err)
			require.Len(t, result.Records, 1)
			require.Equal(t, store.OrphanedRecord{
				ParentId: channel.TeamId,
				ChildId: channel.Id,
			}, result.Records[0])
			close(results)
		})
	})
}
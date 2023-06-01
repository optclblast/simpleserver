package db

import (
	"context"
	"testing"
	"time"

	"server/utils"

	"github.com/stretchr/testify/require"
)

func TestCreateFile(t *testing.T) {
	arg := CreateFileParams{
		Id:          0,
		Owner:       2,
		Name:        "prikil",
		Location:    "/nowhere.mp3",
		LocationWav: "/nowhere.wav",
		LocationTxt: "/nowhere.txt",
		CreatedAt:   time.Now(),
		Status:      "???",
		Guid:        "1234512345",
	}

	file, err := testQueries.CreateFile(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, file)

	require.Equal(t, arg.Id, file.Id)
	require.Equal(t, arg.Owner, file.Owner)
	require.Equal(t, arg.Name, file.Name)
	require.Equal(t, arg.Location, file.Location)
	require.Equal(t, arg.LocationWav, file.LocationWav)
	require.Equal(t, arg.LocationTxt, file.LocationTxt)
	require.Equal(t, arg.Status, file.Status)

	require.NotZero(t, file.Id)
	require.NotZero(t, file.CreatedAt)
}

func TestGetFile(t *testing.T) {
	var id int64 = 178
	var owner int64 = 643
	file, err := testQueries.GetFile(context.Background(), id, owner)
	require.NoError(t, err)
	require.NotEmpty(t, file)

	require.Equal(t, id, file.Id)
	require.Equal(t, "FIELIK", file.Name)
	require.Equal(t, owner, file.Owner)
}

func TestListFiles(t *testing.T) {
	arg := ListFilesParams{
		Limit:  int32(utils.RandomInt(1, 3)),
		Offset: int32(utils.RandomInt(1, 2)),
	}
	files, err := testQueries.ListFiles(context.Background(), arg, 643)
	require.NoError(t, err)
	require.NotEmpty(t, files)
}

func TestUpdateFile(t *testing.T) {
	arg := UpdateFileParams{
		Id:     178,
		Name:   "FIELIKsdf",
		Status: "FFF",
	}
	file, err := testQueries.UpdateFile(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, file)

	require.Equal(t, arg.Id, file.Id)
	require.Equal(t, arg.Name, file.Name)
	require.Equal(t, arg.Status, file.Status)

}

func TestDeleteFile(t *testing.T) {
	err := testQueries.DeleteFile(context.Background(), 8)
	require.NoError(t, err)
}

func TestGetLastFile(t *testing.T) {
	file, err := testQueries.GetLastFileID(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, file)
	var id int64 = 4
	require.Equal(t, id, file)
}
func TestGetFileByGuid(t *testing.T) {
	file, err := testQueries.GetFileByGuid(context.Background(), "fk79995994416_in_2023_05_10-11_46_19_79131488599_zc4z.wav")
	require.NoError(t, err)
	require.NotEmpty(t, file)
	var id int64 = 5
	require.Equal(t, id, file.Id)
}

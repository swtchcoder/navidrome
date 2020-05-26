import React, { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import {
  useCreate,
  useDataProvider,
  useTranslate,
  useNotify,
} from 'react-admin'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from '@material-ui/core'
import { closeAddToPlaylist } from './dialogState'
import SelectPlaylistInput from './SelectPlaylistInput'

const AddToPlaylistDialog = () => {
  const { open, albumId, selectedIds, onSuccess } = useSelector(
    (state) => state.addToPlaylistDialog
  )
  const dispatch = useDispatch()
  const translate = useTranslate()
  const notify = useNotify()
  const [value, setValue] = useState({})
  const dataProvider = useDataProvider()
  const [createAndAddToPlaylist] = useCreate(
    'playlist',
    { name: value.name },
    {
      onSuccess: ({ data }) => {
        setValue(data)
        addToPlaylist(data.id)
      },
      onFailure: (error) => notify(`Error: ${error.message}`, 'warning'),
    }
  )

  const addTracksToPlaylist = (selectedIds, playlistId) =>
    dataProvider
      .create('playlistTrack', {
        data: { ids: selectedIds },
        filter: { playlist_id: playlistId },
      })
      .then(() => selectedIds.length)

  const addAlbumToPlaylist = (albumId, playlistId) =>
    dataProvider
      .getList('albumSong', {
        pagination: { page: 1, perPage: -1 },
        sort: { field: 'discNumber asc, trackNumber asc', order: 'ASC' },
        filter: { album_id: albumId },
      })
      .then((response) => response.data.map((song) => song.id))
      .then((ids) => addTracksToPlaylist(ids, playlistId))

  const addToPlaylist = (playlistId) => {
    const add = albumId
      ? addAlbumToPlaylist(albumId, playlistId)
      : addTracksToPlaylist(selectedIds, playlistId)

    add
      .then((len) => {
        notify('message.songsAddedToPlaylist', 'info', { smart_count: len })
        onSuccess && onSuccess(value, len)
      })
      .catch(() => {
        notify('ra.page.error', 'warning')
      })
  }

  const handleSubmit = (e) => {
    if (value.id) {
      addToPlaylist(value.id)
    } else {
      createAndAddToPlaylist()
    }
    dispatch(closeAddToPlaylist())
    e.stopPropagation()
  }

  const handleClickClose = (e) => {
    dispatch(closeAddToPlaylist())
    e.stopPropagation()
  }

  const handleChange = (pls) => {
    setValue(pls)
  }

  return (
    <Dialog
      open={open}
      onClose={handleClickClose}
      onBackdropClick={handleClickClose}
      aria-labelledby="form-dialog-new-playlist"
    >
      <DialogTitle id="form-dialog-new-playlist">
        {translate('resources.playlist.actions.selectPlaylist')}
      </DialogTitle>
      <DialogContent>
        <SelectPlaylistInput onChange={handleChange} />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClickClose} color="primary">
          {translate('ra.action.cancel')}
        </Button>
        <Button onClick={handleSubmit} color="primary" disabled={!value.name}>
          {translate('ra.action.add')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default AddToPlaylistDialog

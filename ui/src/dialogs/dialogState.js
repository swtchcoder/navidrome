const ADD_TO_PLAYLIST_OPEN = 'ADD_TO_PLAYLIST_OPEN'
const ADD_TO_PLAYLIST_CLOSE = 'ADD_TO_PLAYLIST_CLOSE'

const openAddToPlaylist = ({ albumId, selectedIds, onSuccess }) => ({
  type: ADD_TO_PLAYLIST_OPEN,
  albumId,
  selectedIds,
  onSuccess,
})

const closeAddToPlaylist = () => ({
  type: ADD_TO_PLAYLIST_CLOSE,
})

const addToPlaylistDialogReducer = (
  previousState = {
    open: false,
  },
  payload
) => {
  const { type } = payload
  switch (type) {
    case ADD_TO_PLAYLIST_OPEN:
      return {
        ...previousState,
        open: true,
        albumId: payload.albumId,
        selectedIds: payload.selectedIds,
        onSuccess: payload.onSuccess,
      }
    case ADD_TO_PLAYLIST_CLOSE:
      return { ...previousState, open: false, onSuccess: undefined }
    default:
      return previousState
  }
}

export { openAddToPlaylist, closeAddToPlaylist, addToPlaylistDialogReducer }

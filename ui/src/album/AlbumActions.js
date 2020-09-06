import React from 'react'
import { useDispatch } from 'react-redux'
import {
  Button,
  sanitizeListRestProps,
  TopToolbar,
  useTranslate,
} from 'react-admin'
import PlayArrowIcon from '@material-ui/icons/PlayArrow'
import ShuffleIcon from '@material-ui/icons/Shuffle'
import CloudDownloadOutlinedIcon from '@material-ui/icons/CloudDownloadOutlined'
import AddToQueueIcon from '@material-ui/icons/AddToQueue'
import { addTracks, playTracks, shuffleTracks } from '../audioplayer'
import subsonic from '../subsonic'

const AlbumActions = ({
  albumId,
  className,
  ids,
  data,
  exporter,
  permanentFilter,
  ...rest
}) => {
  const dispatch = useDispatch()
  const translate = useTranslate()

  return (
    <TopToolbar className={className} {...sanitizeListRestProps(rest)}>
      <Button
        onClick={() => {
          dispatch(playTracks(data, ids))
        }}
        label={translate('resources.album.actions.playAll')}
      >
        <PlayArrowIcon />
      </Button>
      <Button
        onClick={() => {
          dispatch(shuffleTracks(data, ids))
        }}
        label={translate('resources.album.actions.shuffle')}
      >
        <ShuffleIcon />
      </Button>
      <Button
        onClick={() => {
          subsonic.download(albumId)
        }}
        label={translate('resources.album.actions.download')}
      >
        <CloudDownloadOutlinedIcon />
      </Button>
      <Button
        onClick={() => {
          dispatch(addTracks(data, ids))
        }}
        label={translate('resources.album.actions.addToQueue')}
      >
        <AddToQueueIcon />
      </Button>
    </TopToolbar>
  )
}

AlbumActions.defaultProps = {
  selectedIds: [],
  onUnselectItems: () => null,
}

export default AlbumActions

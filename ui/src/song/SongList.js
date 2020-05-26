import React from 'react'
import {
  Filter,
  FunctionField,
  NumberField,
  SearchInput,
  TextField,
} from 'react-admin'
import { useMediaQuery } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import StarBorderIcon from '@material-ui/icons/StarBorder'
import StarIcon from '@material-ui/icons/Star'
import {
  DurationField,
  List,
  SimpleList,
  SongContextMenu,
  SongDatagrid,
  SongDetails,
  QuickFilter,
} from '../common'
import { useDispatch } from 'react-redux'
import { setTrack } from '../audioplayer'
import { SongBulkActions } from './SongBulkActions'
import { AlbumLinkField } from './AlbumLinkField'
import AddToPlaylistDialog from '../dialogs/AddToPlaylistDialog'

const useStyles = makeStyles({
  columnIcon: {
    marginLeft: '3px',
    verticalAlign: 'text-top',
  },
})

const SongFilter = (props) => (
  <Filter {...props}>
    <SearchInput source="title" alwaysOn />
    <QuickFilter
      source="starred"
      label={<StarIcon fontSize={'small'} />}
      defaultValue={true}
    />
  </Filter>
)

const SongList = (props) => {
  const classes = useStyles()
  const dispatch = useDispatch()
  const isXsmall = useMediaQuery((theme) => theme.breakpoints.down('xs'))
  const isDesktop = useMediaQuery((theme) => theme.breakpoints.up('md'))
  return (
    <>
      <List
        {...props}
        sort={{ field: 'title', order: 'ASC' }}
        exporter={false}
        bulkActionButtons={<SongBulkActions />}
        filters={<SongFilter />}
        perPage={isXsmall ? 50 : 15}
      >
        {isXsmall ? (
          <SimpleList
            primaryText={(r) => r.title}
            secondaryText={(r) => r.artist}
            tertiaryText={(r) => (
              <>
                <DurationField record={r} source={'duration'} />
                &nbsp;&nbsp;&nbsp;
              </>
            )}
            linkType={(id, basePath, record) => dispatch(setTrack(record))}
            rightIcon={(r) => <SongContextMenu record={r} visible={true} />}
          />
        ) : (
          <SongDatagrid
            expand={<SongDetails />}
            rowClick={(id, basePath, record) => dispatch(setTrack(record))}
          >
            <TextField source="title" />
            {isDesktop && <AlbumLinkField source="album" />}
            <TextField source="artist" />
            {isDesktop && <NumberField source="trackNumber" />}
            {isDesktop && <NumberField source="playCount" />}
            {isDesktop && (
              <FunctionField source="year" render={(r) => r.year || ''} />
            )}
            <DurationField source="duration" />
            {isDesktop ? (
              <SongContextMenu
                source={'starred'}
                label={
                  <StarBorderIcon
                    fontSize={'small'}
                    className={classes.columnIcon}
                  />
                }
                sortBy={'starred DESC, starredAt DESC'}
              />
            ) : (
              <SongContextMenu showStar={false} />
            )}
          </SongDatagrid>
        )}
      </List>
      <AddToPlaylistDialog />,
    </>
  )
}

export default SongList

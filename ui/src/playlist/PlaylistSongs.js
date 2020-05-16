import React from 'react'
import {
  BulkActionsToolbar,
  Datagrid,
  DatagridLoading,
  ListToolbar,
  TextField,
  useListController,
} from 'react-admin'
import classnames from 'classnames'
import { Card, useMediaQuery } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { DurationField, SongDetails } from '../common'

const useStyles = makeStyles(
  (theme) => ({
    root: {},
    main: {
      display: 'flex',
    },
    content: {
      marginTop: 0,
      transition: theme.transitions.create('margin-top'),
      position: 'relative',
      flex: '1 1 auto',
      [theme.breakpoints.down('xs')]: {
        boxShadow: 'none',
      },
    },
    bulkActionsDisplayed: {
      marginTop: -theme.spacing(8),
      transition: theme.transitions.create('margin-top'),
    },
    actions: {
      zIndex: 2,
      display: 'flex',
      justifyContent: 'flex-end',
      flexWrap: 'wrap',
    },
    noResults: { padding: 20 },
  }),
  { name: 'RaList' }
)

const useStylesListToolbar = makeStyles({
  toolbar: {
    justifyContent: 'flex-start',
  },
})

const PlaylistSongs = (props) => {
  const classes = useStyles(props)
  const classesToolbar = useStylesListToolbar(props)
  const isXsmall = useMediaQuery((theme) => theme.breakpoints.down('xs'))
  // const isDesktop = useMediaQuery((theme) => theme.breakpoints.up('md'))
  const controllerProps = useListController(props)
  const { bulkActionButtons, expand, className, playlistId } = props
  const { data, ids, version, loaded } = controllerProps

  const anySong = data[ids[0]]
  const showPlaceholder = !anySong || anySong.playlistId !== playlistId
  const hasBulkActions = props.bulkActionButtons !== false

  if (loaded && ids.length === 0) {
    return <div />
  }

  return (
    <>
      <ListToolbar
        classes={classesToolbar}
        filters={props.filters}
        {...controllerProps}
        actions={props.actions}
        permanentFilter={props.filter}
      />
      <div className={classes.main}>
        <Card
          className={classnames(classes.content, {
            [classes.bulkActionsDisplayed]:
              controllerProps.selectedIds.length > 0,
          })}
          key={version}
        >
          {bulkActionButtons !== false && bulkActionButtons && (
            <BulkActionsToolbar {...controllerProps}>
              {bulkActionButtons}
            </BulkActionsToolbar>
          )}
          {showPlaceholder ? (
            <DatagridLoading
              classes={classes}
              className={className}
              expand={expand}
              hasBulkActions={hasBulkActions}
              nbChildren={3}
              size={'small'}
            />
          ) : (
            <Datagrid
              expand={!isXsmall && <SongDetails />}
              rowClick={null}
              {...controllerProps}
              hasBulkActions={hasBulkActions}
            >
              <TextField source="title" sortable={false} />
              <TextField source="artist" sortable={false} />
              <DurationField source="duration" sortable={false} />
            </Datagrid>
          )}
        </Card>
      </div>
    </>
  )
}

export default PlaylistSongs

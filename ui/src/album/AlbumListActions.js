import React, { cloneElement } from 'react'
import { Button, sanitizeListRestProps, TopToolbar } from 'react-admin'
import { ButtonGroup } from '@material-ui/core'
import ViewHeadlineIcon from '@material-ui/icons/ViewHeadline'
import ViewModuleIcon from '@material-ui/icons/ViewModule'
import { useDispatch, useSelector } from 'react-redux'
import { albumViewGrid, albumViewList } from '../actions'

const AlbumListActions = ({
  currentSort,
  className,
  resource,
  filters,
  displayedFilters,
  filterValues,
  permanentFilter,
  exporter,
  basePath,
  selectedIds,
  onUnselectItems,
  showFilter,
  maxResults,
  total,
  fullWidth,
  ...rest
}) => {
  const dispatch = useDispatch()
  const albumView = useSelector((state) => state.albumView)

  return (
    <TopToolbar className={className} {...sanitizeListRestProps(rest)}>
      {filters &&
        cloneElement(filters, {
          resource,
          showFilter,
          displayedFilters,
          filterValues,
          context: 'button',
        })}
      <ButtonGroup
        variant="text"
        color="primary"
        aria-label="text primary button group"
      >
        <Button
          size="small"
          color={albumView.grid ? 'primary' : 'secondary'}
          onClick={() => dispatch(albumViewGrid())}
        >
          <ViewModuleIcon fontSize="inherit" />
        </Button>
        <Button
          size="small"
          color={albumView.grid ? 'secondary' : 'primary'}
          onClick={() => dispatch(albumViewList())}
        >
          <ViewHeadlineIcon fontSize="inherit" />
        </Button>
      </ButtonGroup>
    </TopToolbar>
  )
}

AlbumListActions.defaultProps = {
  selectedIds: [],
  onUnselectItems: () => null,
}

export default AlbumListActions

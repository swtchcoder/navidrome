import React, { useState, createElement } from 'react'
import { useSelector } from 'react-redux'
import { useMediaQuery } from '@material-ui/core'
import { useTranslate, MenuItemLink, getResources } from 'react-admin'
import { withRouter } from 'react-router-dom'
import LibraryMusicIcon from '@material-ui/icons/LibraryMusic'
import SettingsIcon from '@material-ui/icons/Settings'
import ViewListIcon from '@material-ui/icons/ViewList'
import AlbumIcon from '@material-ui/icons/Album'
import SubMenu from './SubMenu'
import inflection from 'inflection'
import PersonalMenu from './PersonalMenu'
import albumLists from '../album/albumLists'

const translatedResourceName = (resource, translate) =>
  translate(`resources.${resource.name}.name`, {
    smart_count: 2,
    _:
      resource.options && resource.options.label
        ? translate(resource.options.label, {
            smart_count: 2,
            _: resource.options.label,
          })
        : inflection.humanize(inflection.pluralize(resource.name)),
  })

const Menu = ({ onMenuClick, dense, logout }) => {
  const isXsmall = useMediaQuery((theme) => theme.breakpoints.down('xs'))
  const open = useSelector((state) => state.admin.ui.sidebarOpen)
  const translate = useTranslate()
  const resources = useSelector(getResources)

  // TODO State is not persisted in mobile when you close the sidebar menu. Move to redux?
  const [state, setState] = useState({
    menuAlbumList: true,
    menuLibrary: true,
    menuSettings: false,
  })

  const handleToggle = (menu) => {
    setState((state) => ({ ...state, [menu]: !state[menu] }))
  }

  const renderResourceMenuItemLink = (resource) => (
    <MenuItemLink
      key={resource.name}
      to={`/${resource.name}`}
      primaryText={translatedResourceName(resource, translate)}
      leftIcon={
        (resource.icon && createElement(resource.icon)) || <ViewListIcon />
      }
      onClick={onMenuClick}
      sidebarIsOpen={open}
      dense={dense}
    />
  )

  const renderAlbumMenuItemLink = (type, al) => {
    const resource = resources.find((r) => r.name === 'album')
    if (!resource) {
      return null
    }

    const albumListAddress = `/album/${type}`

    const name = translate(`resources.album.lists.${type || 'default'}`, {
      _: translatedResourceName(resource, translate),
    })

    return (
      <MenuItemLink
        key={albumListAddress}
        to={albumListAddress}
        primaryText={name}
        leftIcon={(al.icon && createElement(al.icon)) || <ViewListIcon />}
        onClick={onMenuClick}
        sidebarIsOpen={open}
        dense={dense}
        exact
      />
    )
  }

  const subItems = (subMenu) => (resource) =>
    resource.hasList && resource.options && resource.options.subMenu === subMenu

  return (
    <div>
      <SubMenu
        handleToggle={() => handleToggle('menuAlbumList')}
        isOpen={state.menuAlbumList}
        sidebarIsOpen={open}
        name="menu.albumList"
        icon={<AlbumIcon />}
        dense={dense}
      >
        {Object.keys(albumLists).map((type) =>
          renderAlbumMenuItemLink(type, albumLists[type])
        )}
      </SubMenu>
      <SubMenu
        handleToggle={() => handleToggle('menuLibrary')}
        isOpen={state.menuLibrary}
        sidebarIsOpen={open}
        name="menu.library"
        icon={<LibraryMusicIcon />}
        dense={dense}
      >
        {resources.filter(subItems('library')).map(renderResourceMenuItemLink)}
      </SubMenu>
      <SubMenu
        handleToggle={() => handleToggle('menuSettings')}
        isOpen={state.menuSettings}
        sidebarIsOpen={open}
        name="menu.settings"
        icon={<SettingsIcon />}
        dense={dense}
      >
        {resources.filter(subItems('settings')).map(renderResourceMenuItemLink)}
        <PersonalMenu
          dense={dense}
          sidebarIsOpen={open}
          onClick={onMenuClick}
        />
      </SubMenu>
      {resources.filter(subItems(undefined)).map(renderResourceMenuItemLink)}
      {isXsmall && logout}
    </div>
  )
}

export default withRouter(Menu)

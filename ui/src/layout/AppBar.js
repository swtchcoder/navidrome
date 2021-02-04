import React, { createElement, forwardRef } from 'react'
import {
  AppBar as RAAppBar,
  MenuItemLink,
  useTranslate,
  usePermissions,
  getResources,
} from 'react-admin'
import { useSelector } from 'react-redux'
import { makeStyles, MenuItem, ListItemIcon, Divider } from '@material-ui/core'
import ViewListIcon from '@material-ui/icons/ViewList'
import InfoIcon from '@material-ui/icons/Info'
import { AboutDialog } from '../dialogs'
import PersonalMenu from './PersonalMenu'
import ActivityPanel from './ActivityPanel'
import UserMenu from './UserMenu'
import config from '../config'

const useStyles = makeStyles((theme) => ({
  root: {
    color: theme.palette.text.secondary,
  },
  active: {
    color: theme.palette.text.primary,
  },
  icon: { minWidth: theme.spacing(5) },
}))

const AboutMenuItem = forwardRef(({ onClick, ...rest }, ref) => {
  const classes = useStyles(rest)
  const translate = useTranslate()
  const [open, setOpen] = React.useState(false)

  const handleOpen = () => {
    setOpen(true)
  }
  const handleClose = () => {
    onClick && onClick()
    setOpen(false)
  }
  const label = translate('menu.about')
  return (
    <>
      <MenuItem ref={ref} onClick={handleOpen} className={classes.root}>
        <ListItemIcon className={classes.icon}>
          <InfoIcon titleAccess={label} />
        </ListItemIcon>
        {label}
      </MenuItem>
      <AboutDialog onClose={handleClose} open={open} />
    </>
  )
})

const settingsResources = (resource) =>
  resource.hasList &&
  resource.options &&
  resource.options.subMenu === 'settings'

const CustomUserMenu = ({ onClick, ...rest }) => {
  const translate = useTranslate()
  const resources = useSelector(getResources)
  const classes = useStyles(rest)
  const { permissions } = usePermissions()

  const renderSettingsMenuItemLink = (resource) => {
    const label = translate(`resources.${resource.name}.name`, {
      smart_count: 2,
    })
    return (
      <MenuItemLink
        className={classes.root}
        activeClassName={classes.active}
        key={resource.name}
        to={`/${resource.name}`}
        primaryText={label}
        leftIcon={
          (resource.icon && createElement(resource.icon)) || <ViewListIcon />
        }
        onClick={onClick}
        sidebarIsOpen={true}
      />
    )
  }

  return (
    <>
      {config.devActivityPanel && permissions === 'admin' && <ActivityPanel />}
      <UserMenu {...rest}>
        <PersonalMenu sidebarIsOpen={true} onClick={onClick} />
        <Divider />
        {resources.filter(settingsResources).map(renderSettingsMenuItemLink)}
        <Divider />
        <AboutMenuItem />
      </UserMenu>
    </>
  )
}

const AppBar = (props) => <RAAppBar {...props} userMenu={<CustomUserMenu />} />

export default AppBar

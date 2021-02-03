import React, { useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Layout, toggleSidebar } from 'react-admin'
import { makeStyles } from '@material-ui/core/styles'
import { HotKeys } from 'react-hotkeys'
import Menu from './Menu'
import AppBar from './AppBar'
import Notification from './Notification'
import themes from '../themes'

const useStyles = makeStyles({
  root: { paddingBottom: (props) => (props.addPadding ? '80px' : 0) },
})

export default (props) => {
  const theme = useSelector((state) => themes[state.theme] || themes.DarkTheme)
  const queue = useSelector((state) => state.queue)
  const classes = useStyles({ addPadding: queue.queue.length > 0 })
  const dispatch = useDispatch()

  const keyHandlers = {
    TOGGLE_MENU: useCallback(() => dispatch(toggleSidebar()), [dispatch]),
  }

  return (
    <HotKeys handlers={keyHandlers}>
      <Layout
        {...props}
        className={classes.root}
        menu={Menu}
        appBar={AppBar}
        theme={theme}
        notification={Notification}
      />
    </HotKeys>
  )
}

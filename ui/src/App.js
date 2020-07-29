import React from 'react'
import ReactGA from 'react-ga'
import { Provider } from 'react-redux'
import { createHashHistory } from 'history'
import { Admin, Resource } from 'react-admin'
import dataProvider from './dataProvider'
import authProvider from './authProvider'
import { Layout, Login, Logout } from './layout'
import transcoding from './transcoding'
import player from './player'
import user from './user'
import song from './song'
import album from './album'
import artist from './artist'
import playlist from './playlist'
import { Player, playQueueReducer } from './audioplayer'
import { albumViewReducer } from './album/albumState'
import customRoutes from './routes'
import themeReducer from './personal/themeReducer'
import { addToPlaylistDialogReducer } from './dialogs/dialogState'
import createAdminStore from './store/createAdminStore'
import { i18nProvider } from './i18n'
import config from './config'
const history = createHashHistory()

if (config.gaTrackingId) {
  ReactGA.initialize(config.gaTrackingId)
  history.listen((location) => {
    ReactGA.pageview(location.pathname)
  })
  ReactGA.pageview(window.location.pathname)
}

const App = () => (
  <Provider
    store={createAdminStore({
      authProvider,
      dataProvider,
      history,
      customReducers: {
        queue: playQueueReducer,
        albumView: albumViewReducer,
        theme: themeReducer,
        addToPlaylistDialog: addToPlaylistDialogReducer,
      },
    })}
  >
    <Admin
      dataProvider={dataProvider}
      authProvider={authProvider}
      i18nProvider={i18nProvider}
      customRoutes={customRoutes}
      history={history}
      layout={Layout}
      loginPage={Login}
      logoutButton={Logout}
    >
      {(permissions) => [
        <Resource name="album" {...album} options={{ subMenu: 'albumList' }} />,
        <Resource name="artist" {...artist} options={{ subMenu: 'library' }} />,
        <Resource name="song" {...song} options={{ subMenu: 'library' }} />,
        <Resource
          name="playlist"
          {...playlist}
          options={{ subMenu: 'library' }}
        />,
        permissions === 'admin' ? (
          <Resource name="user" {...user} options={{ subMenu: 'settings' }} />
        ) : null,
        <Resource
          name="player"
          {...player}
          options={{ subMenu: 'settings' }}
        />,
        permissions === 'admin' ? (
          <Resource
            name="transcoding"
            {...transcoding}
            options={{ subMenu: 'settings' }}
          />
        ) : (
          <Resource name="transcoding" />
        ),
        <Resource name="albumSong" />,
        <Resource name="translation" />,
        <Resource name="playlistTrack" />,

        <Player />,
      ]}
    </Admin>
  </Provider>
)

export default App

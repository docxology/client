import * as ConfigGen from '../actions/config-gen'
import * as Styles from '../styles'
import * as DeeplinksGen from '../actions/deeplinks-gen'
import * as React from 'react'
import Main from './main.native'
import configureStore from '../store/configure-store'
import {AppRegistry, AppState, Appearance, Linking} from 'react-native'
import {PortalProvider} from '@gorhom/portal'
import {Provider, useDispatch} from 'react-redux'
import {SafeAreaProvider} from 'react-native-safe-area-context'
import {makeEngine} from '../engine'
import {GestureHandlerRootView} from 'react-native-gesture-handler'

type ConfigureStore = ReturnType<typeof configureStore>
let _store: ConfigureStore | undefined

module.hot?.accept(() => {
  console.log('accepted update in shared/index.native')
})

const NativeEventsToRedux = () => {
  const dispatch = useDispatch()
  const appStateRef = React.useRef('unknown')

  React.useEffect(() => {
    const appStateChangeSub = AppState.addEventListener('change', nextAppState => {
      appStateRef.current = nextAppState
      nextAppState !== 'unknown' &&
        nextAppState !== 'extension' &&
        dispatch(ConfigGen.createMobileAppState({nextAppState}))

      if (nextAppState === 'active') {
        dispatch(ConfigGen.createSetSystemDarkMode({dark: Appearance.getColorScheme() === 'dark'}))
      }
    })

    // only watch dark changes if in foreground due to ios calling this to take snapshots
    const darkSub = Appearance.addChangeListener(() => {
      if (appStateRef.current === 'active') {
        dispatch(ConfigGen.createSetSystemDarkMode({dark: Appearance.getColorScheme() === 'dark'}))
      }
    })
    const linkingSub = Linking.addEventListener('url', ({url}: {url: string}) => {
      dispatch(DeeplinksGen.createLink({link: url}))
    })

    return () => {
      appStateChangeSub.remove()
      darkSub.remove()
      linkingSub.remove()
    }
  }, [dispatch])

  return null
}

const ensureStore = () => {
  if (_store) {
    return
  }
  _store = configureStore()
  if (__DEV__) {
    global.DEBUGStore = _store
  }

  const eng = makeEngine(_store.store.dispatch)
  _store.runSagas()
  eng.sagasAreReady()

  // On mobile there is no installer
  _store.store.dispatch(ConfigGen.createInstallerRan())
}

// on android this can be recreated a bunch so our engine/store / etc should live outside
const Keybase = () => {
  ensureStore()
  if (!_store) return null // never happens
  return (
    <GestureHandlerRootView style={styles.gesture}>
      <Provider store={_store.store}>
        <PortalProvider>
          <SafeAreaProvider>
            <Styles.StyleContext.Provider value={{canFixOverdraw: true}}>
              <Main />
            </Styles.StyleContext.Provider>
          </SafeAreaProvider>
        </PortalProvider>
        <NativeEventsToRedux />
      </Provider>
    </GestureHandlerRootView>
  )
}

const styles = Styles.styleSheetCreate(() => ({
  gesture: {flexGrow: 1},
}))

function load() {
  AppRegistry.registerComponent('Keybase', () => Keybase)
}

export {load}

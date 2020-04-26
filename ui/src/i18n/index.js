import deepmerge from 'deepmerge'
import en from './en'
import cn from './cn'
import fr from './fr'
import it from './it'
import nl from './nl'
import pt from './pt'

const addLanguages = (lang) => {
  Object.keys(lang).forEach((l) => (languages[l] = deepmerge(en, lang[l])))
}
const languages = { en }

// Add new languages to the object bellow (please keep alphabetic sort)
addLanguages({ cn, fr, it, nl, pt })

// "Hack" to make "albumSongs" resource use the same translations as "song"
Object.keys(languages).forEach(
  (k) => (languages[k].resources.albumSong = languages[k].resources.song)
)

export default languages

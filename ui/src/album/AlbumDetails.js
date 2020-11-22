import React, { useMemo } from 'react'
import {
  Card,
  CardContent,
  CardMedia,
  Typography,
  Collapse,
  makeStyles,
  IconButton,
  useMediaQuery,
} from '@material-ui/core'
import classnames from 'classnames'
import { useTranslate } from 'react-admin'
import Lightbox from 'react-image-lightbox'
import 'react-image-lightbox/style.css'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import subsonic from '../subsonic'
import {
  DurationField,
  StarButton,
  SizeField,
  ArtistLinkField,
  formatRange,
  MultiLineTextField,
} from '../common'

const useStyles = makeStyles((theme) => ({
  root: {
    [theme.breakpoints.down('xs')]: {
      padding: '0.7em',
      minWidth: '24em',
    },
    [theme.breakpoints.up('sm')]: {
      padding: '1em',
      minWidth: '32em',
    },
  },
  cardContents: {
    display: 'flex',
  },
  details: {
    display: 'flex',
    flexDirection: 'column',
  },
  content: {
    flex: '2 0 auto',
  },
  coverParent: {
    [theme.breakpoints.down('xs')]: {
      height: '8em',
      width: '8em',
      minWidth: '8em',
    },
    [theme.breakpoints.up('sm')]: {
      height: '10em',
      width: '10em',
      minWidth: '10em',
    },
    [theme.breakpoints.up('lg')]: {
      height: '15em',
      width: '15em',
      minWidth: '15em',
    },
  },
  cover: {
    objectFit: 'contain',
    cursor: 'pointer',
    display: 'block',
    width: '100%',
    height: '100%',
  },
  starButton: {
    top: theme.spacing(-0.2),
    left: theme.spacing(0.5),
  },
  comment: {
    whiteSpace: 'nowrap',
    marginTop: '1em',
    display: 'inline-block',
    [theme.breakpoints.down('xs')]: {
      width: '10em',
    },
    [theme.breakpoints.up('sm')]: {
      width: '10em',
    },
    [theme.breakpoints.up('lg')]: {
      width: '10em',
    },
  },
  commentFirstLine: {
    float: 'left',
    marginRight: '5px',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  },
  expand: {
    marginTop: '-5px',
    transform: 'rotate(0deg)',
    marginLeft: 'auto',
    transition: theme.transitions.create('transform', {
      duration: theme.transitions.duration.shortest,
    }),
  },
  expandOpen: {
    transform: 'rotate(180deg)',
  },
}))

const AlbumComment = ({ classes, record, commentNumLines }) => {
  const [expanded, setExpanded] = React.useState(false)

  const handleExpandClick = React.useCallback(() => {
    commentNumLines > 1 && setExpanded(!expanded)
  }, [expanded, setExpanded, commentNumLines])

  return (
    <div className={classes.comment}>
      <div onClick={handleExpandClick}>
        <MultiLineTextField
          record={record}
          source={'comment'}
          maxLines={1}
          className={classes.commentFirstLine}
        />
      </div>
      {commentNumLines > 1 && (
        <IconButton
          size={'small'}
          className={classnames(classes.expand, {
            [classes.expandOpen]: expanded,
          })}
          onClick={handleExpandClick}
          aria-expanded={expanded}
          aria-label="show more"
        >
          <ExpandMoreIcon />
        </IconButton>
      )}
      <Collapse in={expanded} timeout="auto" unmountOnExit>
        <MultiLineTextField
          record={record}
          source={'comment'}
          firstLine={1}
          className={classes.commentFirstLine}
        />
      </Collapse>
    </div>
  )
}

const AlbumDetails = ({ record }) => {
  const isDesktop = useMediaQuery((theme) => theme.breakpoints.up('lg'))
  const classes = useStyles()
  const [isLightboxOpen, setLightboxOpen] = React.useState(false)
  const translate = useTranslate()

  const commentNumLines = useMemo(
    () => record.comment && record.comment.split('\n').length,
    [record]
  )

  const genreYear = (record) => {
    let genreDateLine = []
    if (record.genre) {
      genreDateLine.push(record.genre)
    }
    const year = formatRange(record, 'year')
    if (year) {
      genreDateLine.push(year)
    }
    return genreDateLine.join(' · ')
  }

  const imageUrl = subsonic.url(
    'getCoverArt',
    record.coverArtId || 'not_found',
    { size: 300 }
  )

  const fullImageUrl = subsonic.url(
    'getCoverArt',
    record.coverArtId || 'not_found'
  )

  const handleOpenLightbox = React.useCallback(() => setLightboxOpen(true), [])
  const handleCloseLightbox = React.useCallback(
    () => setLightboxOpen(false),
    []
  )
  return (
    <Card className={classes.root}>
      <div className={classes.cardContents}>
        <div className={classes.coverParent}>
          <CardMedia
            component={'img'}
            src={imageUrl}
            width="400"
            height="400"
            className={classes.cover}
            onClick={handleOpenLightbox}
            title={record.name}
          />
        </div>
        <div className={classes.details}>
          <CardContent className={classes.content}>
            <Typography variant="h5">
              {record.name}
              <StarButton
                className={classes.starButton}
                record={record}
                resource={'album'}
                size={isDesktop ? 'default' : 'small'}
                aria-label="star"
                color="primary"
              />
            </Typography>
            <Typography component="h6">
              <ArtistLinkField record={record} />
            </Typography>
            <Typography component="p">{genreYear(record)}</Typography>
            <Typography component="p">
              {record.songCount}{' '}
              {translate('resources.song.name', {
                smart_count: record.songCount,
              })}
              {' · '} <DurationField record={record} source={'duration'} />{' '}
              {' · '}
              <SizeField record={record} source="size" />
            </Typography>
            {isDesktop && record['comment'] && (
              <AlbumComment
                classes={classes}
                record={record}
                commentNumLines={commentNumLines}
              />
            )}
          </CardContent>
        </div>
      </div>
      {!isDesktop && record['comment'] && (
        <AlbumComment
          classes={classes}
          record={record}
          commentNumLines={commentNumLines}
        />
      )}
      {isLightboxOpen && (
        <Lightbox
          imagePadding={50}
          animationDuration={200}
          imageTitle={record.name}
          mainSrc={fullImageUrl}
          onCloseRequest={handleCloseLightbox}
        />
      )}
    </Card>
  )
}

export default AlbumDetails

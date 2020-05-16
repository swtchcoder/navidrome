import React from 'react'
import PropTypes from 'prop-types'

const DurationField = ({ record = {}, source }) => {
  try {
    return <span>{format(record[source])}</span>
  } catch (e) {
    console.log('Error in DurationField! Record:', record)
    return <span>00:00</span>
  }
}

const format = (d) => {
  const date = new Date(null)
  date.setSeconds(d)
  const fmt = date.toISOString().substr(11, 8)
  return fmt.replace(/^00:/, '')
}

DurationField.propTypes = {
  label: PropTypes.string,
  record: PropTypes.object,
  source: PropTypes.string.isRequired,
}

DurationField.defaultProps = {
  addLabel: true,
}

export default DurationField

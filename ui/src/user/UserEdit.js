import React from 'react'
import { makeStyles } from '@material-ui/core/styles'
import {
  TextInput,
  BooleanInput,
  DateField,
  PasswordInput,
  Edit,
  required,
  email,
  SimpleForm,
  useTranslate,
  Toolbar,
  SaveButton,
} from 'react-admin'
import { Title } from '../common'
import DeleteUserButton from './DeleteUserButton'

const useStyles = makeStyles({
  toolbar: {
    display: 'flex',
    justifyContent: 'space-between',
  },
})

const UserTitle = ({ record }) => {
  const translate = useTranslate()
  const resourceName = translate('resources.user.name', { smart_count: 1 })
  return <Title subTitle={`${resourceName} ${record ? record.name : ''}`} />
}

const UserToolbar = (props) => (
  <Toolbar {...props} classes={useStyles()}>
    <SaveButton />
    <DeleteUserButton />
  </Toolbar>
)

const UserEdit = (props) => (
  <Edit title={<UserTitle />} {...props}>
    <SimpleForm toolbar={<UserToolbar />}>
      <TextInput source="userName" validate={[required()]} />
      <TextInput source="name" validate={[required()]} />
      <TextInput source="email" validate={[email()]} />
      <PasswordInput source="password" validate={[required()]} />
      <BooleanInput source="isAdmin" initialValue={false} />
      <DateField source="lastLoginAt" showTime />
      {/*<DateField source="lastAccessAt" showTime />*/}
      <DateField source="updatedAt" showTime />
      <DateField source="createdAt" showTime />
    </SimpleForm>
  </Edit>
)

export default UserEdit

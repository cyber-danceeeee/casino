import type { H5PopupPayload } from '../../services/h5Api'

export type BuiltinPopupProps = {
  campaign: H5PopupPayload
}

export type BuiltinPopupEmits = {
  close: []
  action: []
}

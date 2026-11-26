const profileImages = import.meta.glob<string>('../assets/images/profile/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const dialogImages = import.meta.glob<string>('../assets/images/dialogs/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const homeImages = import.meta.glob<string>('../assets/images/home/**/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const depositImages = import.meta.glob<string>('../assets/images/deposit/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const registerRewardImages = import.meta.glob<string>('../assets/images/register-reward/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const offerImages = import.meta.glob<string>('../assets/images/offers/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const vipImages = import.meta.glob<string>('../assets/images/vip/*', {
  eager: true,
  import: 'default',
  query: '?url',
})

const readLocalImage = (
  images: Record<string, string>,
  dir: 'profile' | 'dialogs' | 'home' | 'deposit' | 'register-reward' | 'offers' | 'vip',
  name: string,
) => {
  return images[`../assets/images/${dir}/${name}`] || ''
}

export const localProfileImage = (name: string) => readLocalImage(profileImages, 'profile', name)

export const localDialogImage = (name: string) => readLocalImage(dialogImages, 'dialogs', name)

export const localHomeImage = (name: string) => readLocalImage(homeImages, 'home', name)

export const localDepositImage = (name: string) => readLocalImage(depositImages, 'deposit', name)

export const localRegisterRewardImage = (name: string) =>
  readLocalImage(registerRewardImages, 'register-reward', name)

export const localOfferImage = (name: string) => readLocalImage(offerImages, 'offers', name)

export const localVipImage = (name: string) => readLocalImage(vipImages, 'vip', name)

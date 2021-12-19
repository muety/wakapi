let debounceTimeout

PetiteVue.createApp({
    timezone: guessTimezone(),
    username: '',
    email: '',
    avatarUrl: defaultAvatarUrl,
    updateAvatar() {
        if (!avatarUrlTemplate) return
        if (debounceTimeout) {
            clearTimeout(debounceTimeout)
        }
        debounceTimeout = setTimeout(() => {
            let url = avatarUrlTemplate

            if ((url.includes('{username') && !this.username) || (url.includes('{email') && !this.email)) {
                url = defaultAvatarUrl
            } else {
                url = url.replaceAll('{username}', this.username)
                url = url.replaceAll('{email}', this.email)
                url = url.replaceAll('{username_hash}', MD5(this.username))
                url = url.replaceAll('{email_hash}', MD5(this.email))
                url = url.includes('{') ? defaultAvatarUrl : url
            }
            console.log(url)
            this.avatarUrl = url
        }, 500)
    }
}).mount('#signup-page')
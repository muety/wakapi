import {createApp} from 'petite-vue';
import md5Hex from 'md5-hex';

declare const defaultAvatarUrl = 'assets/images/unknown.svg';
declare const avatarUrlTemplate = 'api/avatar/{username_hash}.svg';

let debounceTimeout: number;

function Signup() {
	return {
		timezone: new Intl.DateTimeFormat().resolvedOptions().timeZone,
		username: '',
		email: '',
		avatarUrl: defaultAvatarUrl,
		updateAvatar() {
			if (!avatarUrlTemplate) {
				return;
			}

			if (debounceTimeout) {
				clearTimeout(debounceTimeout);
			}

			debounceTimeout = setTimeout(() => {
				let url = avatarUrlTemplate;

				if ((url.includes('{username') && !this.username) || (url.includes('{email') && !this.email)) {
					url = defaultAvatarUrl;
				} else {
					url = url.replaceAll('{username}', this.username);
					url = url.replaceAll('{email}', this.email);
					url = url.replaceAll('{username_hash}', md5Hex(this.username));
					url = url.replaceAll('{email_hash}', md5Hex(this.email));
					url = url.includes('{') ? defaultAvatarUrl : url;
				}

				console.log(url);
				this.avatarUrl = url;
			}, 500);
		},
	};
}

createApp(Signup).mount('#signup-page');

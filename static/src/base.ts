export function findParentAttribute(element: Element, attributeName: string) {
	if (element.hasAttribute(attributeName)) {
		return element.getAttribute(attributeName);
	}

	if (!element.parentElement?.hasAttributes()) {
		return null;
	}

	return findParentAttribute(element.parentElement, attributeName);
}

export function copyApiKey(event: Event) {
	const element = document.querySelector('input#api-key-container')!;
	element.select();
	element.setSelectionRange(0, 9999);
	document.execCommand('copy');
	event.stopPropagation();
}

window.addEventListener('load', () => {
	const baseUrl = location.href.slice(0, Math.max(0, location.href.lastIndexOf('/')));
	for (const element of document.querySelectorAll('.with-url-src')) {
		element.setAttribute('src', element.getAttribute('src')!.replace('%s', baseUrl));
	}

	for (const element of document.querySelectorAll('.with-url-src-no-scheme')) {
		const strippedUrl = baseUrl.replace(/https?:\/\//, '');
		element.setAttribute('src', element.getAttribute('src')!.replace('%s', strippedUrl));
	}

	for (const element of document.querySelectorAll('.with-url-value')) {
		element.setAttribute('value', element.getAttribute('value')!.replace('%s', baseUrl));
	}

	for (const element of document.querySelectorAll('.with-url-inner')) {
		element.innerHTML = element.innerHTML.replace('%s', baseUrl);
	}

	for (const element of document.querySelectorAll('.with-url-inner-no-scheme')) {
		const strippedUrl = baseUrl.replace(/https?:\/\//, '');
		element.innerHTML = element.innerHTML.replace('%s', strippedUrl);
	}
});

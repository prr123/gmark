const hdMdObj = {
	style: {
		color: 'Green',
		margin: 'auto',
		textAlign: 'center',
        padding: '0.5rem',
		fontSize: '2rem',
    },
	id: 'docmainHd',
	textContent: site.name,
	typ: 'h1',
};

const hdel = azul.addElement(hdMdObj);
azul.docbody.appendChild(hdel);
let mdDiv = site.render();
azul.docbody.appendChild(mdDiv);


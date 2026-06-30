// renderHandoffQR draws the branded onboarding QR — the one with the cross
// (verified.png) in the center and the maroon classy-rounded dots — into the
// given container element. It encodes the user's fresh-login URL so a phone
// camera scan completes the hand-off. Styling matches v1's show_user_uuid_qrcode.
function renderHandoffQR(container, uuid, opts) {
  opts = opts || {};
  var size = opts.size || 300;
  var url = opts.url || (window.location.origin + '/user/login/fresh/' + uuid);
  if (typeof QRCodeStyling === 'undefined') {
    container.textContent = 'QR library failed to load';
    return null;
  }
  var qr = new QRCodeStyling({
    width: size,
    height: size,
    type: 'svg',
    data: url,
    image: '/static/verified.png',
    dotsOptions: { color: '#913C67', type: 'classy-rounded' },
    cornersSquareOptions: { color: '#913C67', type: 'extra-rounded' },
    backgroundOptions: { color: '#e9ebee' },
    imageOptions: { crossOrigin: 'anonymous', margin: 4 }
  });
  container.innerHTML = '';
  qr.append(container);
  var svg = container.querySelector('svg');
  if (svg) svg.classList.add('rounded');
  return qr;
}

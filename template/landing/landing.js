/* =============================================
   i18n.js — Multi-language support
   ============================================= */
var i18n = (function() {
    'use strict';
    var currentLang = 'hy';
    var fallbackLang = 'en';

    var dict = {
        en: {
            'header.support': '24/7 support',
            'hero.title': 'Start Your Free<br>Security Trial',
            'hero.subtitle': 'Get 7 days of full protection — no payment required. Just enter your phone number to get started.',
            'card.title': 'Get Your Free Trial',
            'card.subtitle': 'Enter your phone number and receive the activation code via SMS',
            'form.phoneLabel': 'Phone Number <span>*</span>',
            'form.phonePlaceholder': '+374 XX XXX XXX',
            'form.submitBtn': 'Start Free Trial',
            'form.sendCode': 'Send SMS code',
            'form.confirmOtp': 'Confirm subscription',
            'form.otpLabel': 'Code from SMS <span>*</span>',
            'form.otpPlaceholder': 'Enter code',
            'form.otpError': 'Enter the code from the SMS',
            'form.phoneError': 'Please enter a valid phone number in +374 XX XXX XXX format',
            'result.found': 'Trial Activated',
            'result.notFound': 'Activation Failed',
            'result.status': 'Status',
            'result.product': 'Product',
            'result.start': 'Start Date',
            'result.end': 'End Date',
            'result.daysLeft': 'Days Left',
            'result.devices': 'Devices',
            'result.active': 'Active',
            'result.trial': 'Trial',
            'result.expired': 'Expired',
            'result.cancelled': 'Cancelled',
            'result.successMsg': 'Your free trial is active! Download the app and use the activation code to start protecting your device.',
            'result.errorMsg': 'Could not activate the trial. Please check your number and try again.',
            'result.alreadySubscribed': 'This number already has an active subscription.',
            'result.applicationSubmitted': 'Thank you! Your request has been submitted.',
            'result.requestFailed': 'Something went wrong. Please try again.',
            'result.apiError': 'Error: ',
            'features.security.title': 'Security',
            'features.security.desc': 'Real-time protection from viruses and threats',
            'features.instant.title': 'Instant',
            'features.instant.desc': 'Activation code sent via SMS in seconds',
            'features.always.title': '24/7',
            'features.always.desc': 'Your trial is active around the clock',
            'footer.privacy': 'Privacy Policy',
            'footer.terms': 'Terms of Use',
            'footer.contact': 'Contact Us',
            'footer.copyright': '© {year} Viva Armenia. All rights reserved.'
        },
        ru: {
            'header.support': 'Поддержка 24/7',
            'hero.title': 'Начните бесплатный<br>пробный период',
            'hero.subtitle': '7 дней полной защиты — без оплаты. Просто введите номер телефона, чтобы начать.',
            'card.title': 'Получить пробный период',
            'card.subtitle': 'Введите номер телефона и получите код активации по SMS',
            'form.phoneLabel': 'Номер телефона <span>*</span>',
            'form.phonePlaceholder': '+374 XX XXX XXX',
            'form.submitBtn': 'Начать пробный период',
            'form.sendCode': 'Получить код по SMS',
            'form.confirmOtp': 'Подтвердить подписку',
            'form.otpLabel': 'Код из SMS <span>*</span>',
            'form.otpPlaceholder': 'Введите код',
            'form.otpError': 'Введите код из SMS',
            'form.phoneError': 'Введите корректный номер телефона в формате +374 XX XXX XXX',
            'result.found': 'Пробный период активирован',
            'result.notFound': 'Не удалось активировать',
            'result.status': 'Статус',
            'result.product': 'Продукт',
            'result.start': 'Начало',
            'result.end': 'Окончание',
            'result.daysLeft': 'Осталось дней',
            'result.devices': 'Устройства',
            'result.active': 'Активна',
            'result.trial': 'Пробная',
            'result.expired': 'Истекла',
            'result.cancelled': 'Отменена',
            'result.successMsg': 'Ваш пробный период активен! Скачайте приложение и используйте код активации для защиты устройства.',
            'result.errorMsg': 'Не удалось активировать пробный период. Проверьте номер и попробуйте снова.',
            'result.alreadySubscribed': 'На этот номер уже оформлена подписка.',
            'result.applicationSubmitted': 'Спасибо! Ваша заявка направлена.',
            'result.requestFailed': 'Что-то пошло не так. Попробуйте ещё раз.',
            'result.apiError': 'Ошибка: ',
            'features.security.title': 'Безопасность',
            'features.security.desc': 'Защита от вирусов и угроз в реальном времени',
            'features.instant.title': 'Мгновенно',
            'features.instant.desc': 'Код активации приходит по SMS за секунды',
            'features.always.title': '24/7',
            'features.always.desc': 'Пробный период активен круглосуточно',
            'footer.privacy': 'Политика конфиденциальности',
            'footer.terms': 'Условия использования',
            'footer.contact': 'Связаться с нами',
            'footer.copyright': '© {year} Viva Armenia. Все права защищены.'
        },
        hy: {
            'header.support': '24/7 աջակցություն',
            'hero.title': 'Սկսեք անվճար<br>փորձնական ամկետ',
            'hero.subtitle': '7 օր ամբողջական պաշտպանություն՝ առանց վճարման։ Պարզապես մուտքագրեք հեռախոսահամարը։',
            'card.title': 'Ստանալ փորձնական ամկետ',
            'card.subtitle': 'Մուտքագրեք հեռախոսահամարը և ստացեք ակտիվացման կոդը SMS-ով',
            'form.phoneLabel': 'Հեռախոսահամար <span>*</span>',
            'form.phonePlaceholder': '+374 XX XXX XXX',
            'form.submitBtn': 'Սկսել փորձնականը',
            'form.sendCode': 'Ուղարկել կոդը SMS-ով',
            'form.confirmOtp': 'Հաստատել բաժանորդագրությունը',
            'form.otpLabel': 'Կոդը SMS-ից <span>*</span>',
            'form.otpPlaceholder': 'Մուտքագրեք կոդը',
            'form.otpError': 'Մուտքագրեք SMS-ում ստացած կոդը',
            'form.phoneError': 'Խնդրում ենք մուտքագրել ճիշտ հեռախոսահամար +374 XX XXX XXX ձևաչափով',
            'result.found': 'Փորձնականը ակտիվացված է',
            'result.notFound': 'Ակտիվացումը ձախողվեց',
            'result.status': 'Կարգավիճակ',
            'result.product': 'Արտադրանք',
            'result.start': 'Սկիզբ',
            'result.end': 'Ավարտ',
            'result.daysLeft': 'նացած օրեր',
            'result.devices': 'Սարքեր',
            'result.active': 'Ակտիվ',
            'result.trial': 'Փորձնական',
            'result.expired': 'Ժամկետանց',
            'result.cancelled': 'Չեղարկված',
            'result.successMsg': 'Ձեր փորձնականը ակտիվ է։ Ներբեռնեք հավելվածը և օգտագործեք ակտիվացման կոդը։',
            'result.errorMsg': 'Չհաջողվեց ակտիվացնել փորնականը։ Ստուգեք համարը և փորեք կրկին։',
            'result.alreadySubscribed': 'Այս համարը արդեն ունի ակտիվ բաժանորդագրություն։',
            'result.applicationSubmitted': 'Շնորհակալություն։ Ձեր դիմումն ուղարկված է։',
            'result.requestFailed': 'Ինչ-որ բան սխալ է գնացել։ Փորձեք կրկին։',
            'result.apiError': 'Սխալ։ ',
            'features.security.title': 'Անվտանգություն',
            'features.security.desc': 'Պաշտպանություն վիրուսներից և սպառնալիքներից իրական ժամանակում',
            'features.instant.title': 'Ակնթարթային',
            'features.instant.desc': 'Ակտիվացման կոդը ուղարկվում է SMS-ով վայրկյանների ընթացքում',
            'features.always.title': '24/7',
            'features.always.desc': 'Փորձնականը ակտիվ է շուրջօրյա',
            'footer.privacy': 'Գաղտնիության քաղաքականություն',
            'footer.terms': 'Օգտագործման պայմաններ',
            'footer.contact': 'Կապվել մեզ հետ',
            'footer.copyright': '© {year} Viva Armenia. Բոլոր իրավունքները պաշտպանված են։'
        }
    };

    function t(key, fallback) {
        var langDict = dict[currentLang] || dict[fallbackLang];
        var val = langDict[key];
        if (val === undefined) {
            if (fallback) val = fallback;
            else val = dict[fallbackLang][key] || key;
        }
        if (typeof val === 'string' && val.indexOf('{year}') !== -1) {
            val = val.replace(/\{year\}/g, String(new Date().getFullYear()));
        }
        return val;
    }

    function setLang(lang) {
        if (!dict[lang]) lang = fallbackLang;
        currentLang = lang;
        document.documentElement.lang = lang === 'hy' ? 'hy' : lang;

        document.querySelectorAll('[data-i18n]').forEach(function(el) {
            var key = el.getAttribute('data-i18n');
            el.innerHTML = t(key);
        });

        document.querySelectorAll('[data-i18n-placeholder]').forEach(function(el) {
            var key = el.getAttribute('data-i18n-placeholder');
            el.setAttribute('placeholder', t(key));
        });

        document.querySelectorAll('.viva-lang-btn').forEach(function(btn) {
            btn.classList.toggle('active', btn.getAttribute('data-lang') === lang);
        });

        try { localStorage.setItem('viva_lang', lang); } catch(e) {}
        document.dispatchEvent(new CustomEvent('viva-lang-change', { detail: { lang: lang } }));
    }

    function getLang() { return currentLang; }

    function addTranslations(lang, obj) {
        if (!dict[lang]) dict[lang] = {};
        for (var key in obj) {
            if (obj.hasOwnProperty(key)) dict[lang][key] = obj[key];
        }
    }

    function init() {
        var saved = null;
        try { saved = localStorage.getItem('viva_lang'); } catch(e) {}
        if (saved && dict[saved]) {
            setLang(saved);
        } else {
            setLang('hy');
        }
    }

    return { t: t, setLang: setLang, getLang: getLang, addTranslations: addTranslations, init: init, dict: dict };
})();

/* =============================================
   phone-formatter.js
   ============================================= */
var PhoneFormatter = (function() {
    'use strict';
    return {
        format: function(raw) {
            var digits = raw.replace(/\D/g, '');
            if (!digits.startsWith('374')) {
                if (digits.startsWith('0')) digits = '374' + digits.slice(1);
                else if (!digits.startsWith('3')) digits = '374' + digits;
            }
            var formatted = '+374';
            var rest = digits.slice(3);
            if (rest.length > 0) formatted += ' ' + rest.slice(0, 2);
            if (rest.length > 2) formatted += ' ' + rest.slice(2, 5);
            if (rest.length > 5) formatted += ' ' + rest.slice(5, 8);
            return formatted;
        },
        clean: function(formatted) {
            return '+' + formatted.replace(/\D/g, '');
        },
        validate: function(phone) {
            var digits = phone.replace(/\D/g, '');
            return digits.length >= 11 && digits.length <= 14 && digits.startsWith('374');
        }
    };
})();

/* =============================================
   api-client.js — Landing endpoints
   ============================================= */
var APIClient = (function() {
    'use strict';
    var baseUrl = '';

    function landingPath(rest) {
        var lang = (typeof i18n !== 'undefined' && i18n.getLang) ? i18n.getLang() : 'hy';
        return '/landing/' + lang + rest;
    }

    return {
        setBaseUrl: function(url) { baseUrl = String(url || '').replace(/\/+$/, ''); },
        _request: function(method, path, body) {
            var url = baseUrl + path;
            var opts = { method: method, headers: {} };
            if (body != null && method !== 'GET' && method !== 'HEAD') {
                opts.headers['Content-Type'] = 'application/json';
                opts.body = JSON.stringify(body);
            }
            return fetch(url, opts).then(function(resp) {
                return resp.text().then(function(text) {
                    var json = null;
                    if (text) {
                        try { json = JSON.parse(text); } catch (e) { /* non-JSON body */ }
                    }
                    if (!resp.ok) {
                        var msg = (json && json.error) ? json.error : ('HTTP ' + resp.status + (text ? ': ' + text.slice(0, 400) : ''));
                        throw new Error(msg);
                    }
                    return json || {};
                });
            });
        },
        initSubscription: function(data) {
            return this._request('POST', landingPath('/init-subscription'), data);
        },
        confirmSubscription: function(data) {
            return this._request('POST', landingPath('/confirm-subscription'), data);
        },
        getSubscriberInfo: function(phone) {
            var digits = String(phone).replace(/\D/g, '');
            return this._request('GET', landingPath('/subscriber-info/' + encodeURIComponent(digits)), null);
        },
        postSubscriberInfo: function(data) {
            return this._request('POST', landingPath('/subscriber-info'), data);
        }
    };
})();

/* =============================================
   ui-helpers.js
   ============================================= */
var UI = (function() {
    'use strict';
    function $(sel) { return document.querySelector(sel); }

    function showLoading(btn) { btn.classList.add('viva-btn--loading'); btn.disabled = true; }
    function hideLoading(btn) { btn.classList.remove('viva-btn--loading'); btn.disabled = false; }

    function showError(input, errorEl, msg) {
        input.classList.add('input-error');
        if (errorEl) { errorEl.innerHTML = msg || ''; errorEl.classList.add('show'); }
    }
    function clearError(input, errorEl) {
        input.classList.remove('input-error');
        if (errorEl) errorEl.classList.remove('show');
    }

    function esc(str) {
        var div = document.createElement('div');
        div.textContent = str || '';
        return div.innerHTML;
    }

    function renderResult(container, data) {
        var statusVal = data.status || 'error';
        var iconWrapClass = 'viva-result__icon--error';
        var titleClass = 'viva-result__title--error';
        var cardClass = 'viva-result__card--error';
        var iconSvg =
            '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>';
        var titleText = esc(data.title || i18n.t('result.notFound'));

        if (statusVal === 'submitted') {
            iconWrapClass = 'viva-result__icon--success';
            titleClass = 'viva-result__title--success';
            cardClass = 'viva-result__card--success';
            iconSvg =
                '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>';
            titleText = esc(data.title || i18n.t('result.applicationSubmitted'));
        } else if (statusVal === 'trial' || statusVal === 'info') {
            iconWrapClass = 'viva-result__icon--info';
            titleClass = 'viva-result__title--info';
            cardClass = 'viva-result__card--info';
            iconSvg =
                '<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>';
            titleText = esc(data.title || i18n.t('result.alreadySubscribed'));
        }

        container.innerHTML =
            '<div class="viva-result__card viva-result__card--compact ' + cardClass + '">' +
                '<div class="viva-result__header viva-result__header--standalone">' +
                    '<div class="viva-result__icon ' + iconWrapClass + '">' + iconSvg + '</div>' +
                    '<div class="viva-result__title ' + titleClass + '">' + titleText + '</div>' +
                '</div>' +
            '</div>';
        container.classList.add('show');
    }

    return {
        $: $,
        showLoading: showLoading, hideLoading: hideLoading,
        showError: showError, clearError: clearError,
        renderResult: renderResult, esc: esc
    };
})();

/* =============================================
   app.js — Main application
   ============================================= */
var App = (function() {
    'use strict';

    var form, phoneInput, phoneError, otpStep, otpInput, otpError, submitBtn, submitBtnText, resultBlock;
    var subscriptionStage = 'phone';
    var pendingPhone = '';

    function mapLandingSubscriptionResult(data) {
        var confirm = data.confirm;
        var init = data.init;
        var m = confirm || init;
        if (!m) {
            return { status: 'error', title: i18n.t('result.notFound') };
        }
        var code = m.resultCode;
        if (confirm && (code === 0 || m.result === true)) {
            return { status: 'submitted', title: i18n.t('result.applicationSubmitted') };
        }
        if (code === 7) {
            return { status: 'info', title: i18n.t('result.alreadySubscribed') };
        }
        return { status: 'error', title: i18n.t('result.notFound') };
    }

    function subscriptionPayload(phone, extra) {
        var o = {
            phoneNum: phone,
            locale: i18n.getLang()
        };
        if (extra) {
            for (var k in extra) {
                if (Object.prototype.hasOwnProperty.call(extra, k)) o[k] = extra[k];
            }
        }
        return o;
    }

    function resetOtpStep() {
        subscriptionStage = 'phone';
        pendingPhone = '';
        if (otpStep) {
            otpStep.classList.remove('is-visible');
            otpStep.setAttribute('aria-hidden', 'true');
        }
        if (otpInput) otpInput.value = '';
        if (otpInput && otpError) UI.clearError(otpInput, otpError);
        updateSubmitLabel();
    }

    function updateSubmitLabel() {
        if (!submitBtnText) return;
        if (subscriptionStage === 'otp') {
            submitBtnText.innerHTML = i18n.t('form.confirmOtp');
        } else {
            submitBtnText.innerHTML = i18n.t('form.sendCode');
        }
    }

    function resolveApiBase() {
        var meta = document.querySelector('meta[name="viva-api-base"]');
        if (meta && meta.content && meta.content.trim()) {
            return meta.content.trim();
        }
        var h = window.location.hostname;
        if (h === 'localhost' || h === '127.0.0.1') {
            return window.location.protocol + '//' + h + ':4000';
        }
        return '';
    }

    function init() {
        var base = resolveApiBase();
        if (base) {
            APIClient.setBaseUrl(base);
        }

        form = UI.$('#checkForm');
        phoneInput = UI.$('#phoneInput');
        phoneError = UI.$('#phoneError');
        otpStep = UI.$('#otpStep');
        otpInput = UI.$('#otpInput');
        otpError = UI.$('#otpError');
        submitBtn = UI.$('#submitBtn');
        submitBtnText = submitBtn ? submitBtn.querySelector('.viva-btn__text') : null;
        resultBlock = UI.$('#resultBlock');

        initLangSwitcher();
        bindEvents();
        updateSubmitLabel();

        document.addEventListener('viva-lang-change', function() { updateSubmitLabel(); });
    }

    function initLangSwitcher() {
        document.querySelectorAll('.viva-lang-btn').forEach(function(btn) {
            btn.addEventListener('click', function() {
                i18n.setLang(this.getAttribute('data-lang'));
            });
        });
    }

    function bindEvents() {
        phoneInput.addEventListener('input', function() {
            var val = this.value;
            var formatted = PhoneFormatter.format(val);
            if (formatted !== val) phoneInput.value = formatted;
            UI.clearError(phoneInput, phoneError);
            resetOtpStep();
        });

        phoneInput.addEventListener('paste', function(e) {
            var pasted = (e.clipboardData || window.clipboardData).getData('text');
            e.preventDefault();
            phoneInput.value = PhoneFormatter.format(pasted);
            UI.clearError(phoneInput, phoneError);
            resetOtpStep();
        });

        if (otpInput) {
            otpInput.addEventListener('input', function() {
                UI.clearError(otpInput, otpError);
            });
        }

        form.addEventListener('submit', function(e) {
            e.preventDefault();
            if (subscriptionStage === 'otp') {
                handleConfirmOtp();
            } else {
                handleSubmit();
            }
        });
    }

    function handleSubmit() {
        var phone = phoneInput.value;
        var cleanPhone = PhoneFormatter.clean(phone);

        if (!PhoneFormatter.validate(phone)) {
            UI.showError(phoneInput, phoneError, i18n.t('form.phoneError'));
            phoneInput.focus();
            return;
        }

        UI.showLoading(submitBtn);
        resultBlock.classList.remove('show');

        APIClient.initSubscription(subscriptionPayload(cleanPhone, { skipConfirm: true }))
            .then(function(data) {
                var ir = data.init;
                if (!ir) {
                    UI.renderResult(resultBlock, { status: 'error', title: i18n.t('result.notFound') });
                    return;
                }
                if (ir.resultCode === 7) {
                    UI.renderResult(resultBlock, mapLandingSubscriptionResult({ init: ir }));
                    return;
                }
                var okInit = (ir.resultCode === 0 || ir.result === true);
                if (!okInit) {
                    UI.renderResult(resultBlock, { status: 'error', title: i18n.t('result.notFound') });
                    return;
                }
                pendingPhone = cleanPhone;
                subscriptionStage = 'otp';
                otpStep.classList.add('is-visible');
                otpStep.setAttribute('aria-hidden', 'false');
                updateSubmitLabel();
                otpInput.focus();
            })
            .catch(function() {
                UI.renderResult(resultBlock, { status: 'error', title: i18n.t('result.requestFailed') });
            })
            .finally(function() {
                UI.hideLoading(submitBtn);
            });
    }

    function handleConfirmOtp() {
        var otp = (otpInput.value || '').trim();
        if (!otp) {
            UI.showError(otpInput, otpError, i18n.t('form.otpError'));
            otpInput.focus();
            return;
        }
        if (!pendingPhone) {
            resetOtpStep();
            return;
        }

        UI.showLoading(submitBtn);
        resultBlock.classList.remove('show');

        APIClient.confirmSubscription(subscriptionPayload(pendingPhone, { otp: otp }))
            .then(function(data) {
                var mapped = mapLandingSubscriptionResult(data);
                UI.renderResult(resultBlock, mapped);
                if (mapped.status === 'submitted' && form) {
                    form.classList.add('viva-form--hidden');
                    var card = form.closest('.viva-card');
                    if (card) {
                        var sub = card.querySelector('.viva-card__subtitle');
                        if (sub) sub.classList.add('viva-form--hidden');
                    }
                }
                resetOtpStep();
            })
            .catch(function() {
                UI.renderResult(resultBlock, { status: 'error', title: i18n.t('result.requestFailed') });
            })
            .finally(function() {
                UI.hideLoading(submitBtn);
            });
    }

    return { init: init };
})();

/* ===== Bootstrap ===== */
document.addEventListener('DOMContentLoaded', function() {
    i18n.init();
    App.init();
});

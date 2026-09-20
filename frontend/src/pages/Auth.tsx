import { Component, createSignal, Show } from 'solid-js';
import { useNavigate } from '@solidjs/router';
import { useAuth } from '../store/auth';
import { useI18n } from '../store/i18n';

export const Auth: Component = () => {
  const navigate = useNavigate();
  const { login, register } = useAuth();
  const { t } = useI18n();

  const [isLogin, setIsLogin] = createSignal(true);
  const [fullName, setFullName] = createSignal('');
  const [email, setEmail] = createSignal('');
  const [password, setPassword] = createSignal('');
  const [errorMsg, setErrorMsg] = createSignal('');
  const [loading, setLoading] = createSignal(false);

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    setErrorMsg('');
    setLoading(true);

    try {
      if (isLogin()) {
        await login(email(), password());
      } else {
        await register(fullName(), email(), password());
      }
      navigate('/account');
    } catch (err: any) {
      setErrorMsg(err.message || 'Authenticatiefout');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div class="container" style={{ padding: '4rem 1.5rem 6rem', 'max-width': '460px' }}>
      <div class="glass-card" style={{ padding: '2.5rem 2rem' }}>
        <h2 style={{ "text-align": "center", "font-size": "1.75rem", "margin-bottom": "1.5rem" }}>
          {isLogin() ? t('auth_welcome_back') : t('auth_create_account')}
        </h2>

        <Show when={errorMsg()}>
          <div style={{ color: 'var(--danger)', 'margin-bottom': '1rem', 'font-size': '0.9rem', 'text-align': 'center' }}>
            {errorMsg()}
          </div>
        </Show>

        <form onSubmit={handleSubmit} style={{ display: 'flex', 'flex-direction': 'column', gap: '1rem' }}>
          <Show when={!isLogin()}>
            <div>
              <label style={{ "font-size": "0.8rem", "font-weight": 600 }}>{t('auth_full_name')}</label>
              <input
                type="text"
                required
                class="glass-input"
                value={fullName()}
                onInput={(e) => setFullName(e.currentTarget.value)}
              />
            </div>
          </Show>

          <div>
            <label style={{ "font-size": "0.8rem", "font-weight": 600 }}>{t('auth_email')}</label>
            <input
              type="email"
              required
              class="glass-input"
              value={email()}
              onInput={(e) => setEmail(e.currentTarget.value)}
            />
          </div>

          <div>
            <label style={{ "font-size": "0.8rem", "font-weight": 600 }}>{t('auth_password')}</label>
            <input
              type="password"
              required
              minlength="6"
              class="glass-input"
              value={password()}
              onInput={(e) => setPassword(e.currentTarget.value)}
            />
          </div>

          <button type="submit" class="glass-button" disabled={loading()} style={{ 'margin-top': '0.5rem', height: '48px' }}>
            {loading() ? t('auth_processing') : isLogin() ? t('auth_login') : t('auth_register')}
          </button>
        </form>

        <div style={{ "text-align": "center", "margin-top": "1.5rem", "font-size": "0.9rem" }}>
          <button
            style={{ color: "var(--accent)", "font-weight": 600 }}
            onClick={() => { setIsLogin(!isLogin()); setErrorMsg(''); }}
          >
            {isLogin() ? t('auth_no_account') : t('auth_has_account')}
          </button>
        </div>
      </div>
    </div>
  );
};
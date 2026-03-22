import { createApp } from 'vue';
import App from './App.vue';
import PrimeVue from 'primevue/config';
import Button from 'primevue/button';
import Card from 'primevue/card';
import Chart from 'primevue/chart';
import Tag from 'primevue/tag';
import Badge from 'primevue/badge';

// PrimeVue styles
// Import the Aura preset from @primevue/themes. As of PrimeVue v4, themes are provided
// via the @primevue/themes package instead of primevue/resources/themes.
import Aura from '@primevue/themes/aura';

// PrimeIcons for icons
import 'primeicons/primeicons.css';

// Chart.js auto registration is required for PrimeVue Chart component
import 'chart.js/auto';

// Tailwind styles
import './index.css';

const app = createApp(App);

// Configure PrimeVue. Enable ripple effect and apply the Aura theme preset. The
// theme options can be customized; here we use default values with the CSS
// variables prefixed by "p" and dark mode set to follow system preference.
app.use(PrimeVue, {
  ripple: true,
  theme: {
    preset: Aura,
    options: {
      prefix: 'p',
      darkModeSelector: 'system',
      cssLayer: false
    }
  }
});

// Register commonly used components globally
app.component('Button', Button);
app.component('Card', Card);
app.component('Chart', Chart);
app.component('Tag', Tag);
app.component('Badge', Badge);

app.mount('#app');
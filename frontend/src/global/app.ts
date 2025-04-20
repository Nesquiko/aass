import { registerNavigationApi } from './navigation';
import '@material/web/button/filled-button';
import '@material/web/button/outlined-button';
import '@material/web/button/text-button';
import '@material/web/checkbox/checkbox';
import '@material/web/divider/divider';
import '@material/web/field/filled-field';
import '@material/web/icon/icon';
import '@material/web/iconbutton/icon-button';
import '@material/web/list/list';
import '@material/web/list/list-item';
import '@material/web/menu/menu';
import '@material/web/menu/menu-item';
import '@material/web/ripple/ripple';
import '@material/web/select/filled-select';
import '@material/web/select/outlined-select';
import '@material/web/select/select-option';
import '@material/web/switch/switch';
import '@material/web/tabs/primary-tab';
import '@material/web/tabs/secondary-tab';
import '@material/web/tabs/tabs';
import '@material/web/textfield/filled-text-field';
import '@material/web/textfield/outlined-text-field';

export default function () {
  registerNavigationApi();
}

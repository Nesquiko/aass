import { Api, newApi } from '../../api/api';
import { User } from '../../api/generated';
import { StyledHost } from '../StyledHost';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-app',
})
export class App {
  @Prop() apiBase: string;
  @Prop() basePath: string = '';

  @State() private relativePath: string = '';

  private api: Api;

  constructor() {
    this.api = newApi(this.apiBase);
  }

  componentWillLoad() {
    const baseUri = new URL(this.basePath, document.baseURI || '/').pathname;

    const toRelative = (path: string) => {
      if (path.startsWith(baseUri)) {
        this.relativePath = path.slice(baseUri.length);
      } else {
        this.relativePath = '';
      }
    };

    window.navigation?.addEventListener('navigate', (ev: NavigateEvent) => {
      if (ev.canIntercept) {
        ev.intercept();
      }
      const path = new URL(ev.destination.url).pathname;
      toRelative(path);
    });

    toRelative(location.pathname);
  }

  render() {
    let element: Element;
    const user: User | null = JSON.parse(sessionStorage.getItem('user'));

    const navigate = (path: string) => {
      const absolute = new URL(path, new URL(this.basePath, document.baseURI));
      window.navigation.navigate(absolute.pathname + absolute.search);
    };

    if (!user && !this.relativePath.startsWith('register')) {
      element = <xcastven-xkilian-project-login api={this.api} navigate={navigate} />;
    } else {
      if (this.relativePath === '' || this.relativePath.startsWith('homepage')) {
        element = <xcastven-xkilian-project-home-page navigate={navigate} api={this.api} />;
      } else if (this.relativePath.startsWith('login')) {
        element = <xcastven-xkilian-project-login api={this.api} navigate={navigate} />;
      } else if (this.relativePath.startsWith('registerCondition')) {
        const urlParams = new URLSearchParams(window.location.search);
        const start: string = urlParams.get('start');
        element = (
          <xcastven-xkilian-project-condition-registerer
            navigate={navigate}
            api={this.api}
            user={user}
            startDate={start ? new Date(start) : null}
          />
        );
      } else if (this.relativePath.startsWith('register')) {
        element = <xcastven-xkilian-project-register navigate={navigate} api={this.api} />;
      } else if (this.relativePath.startsWith('scheduleAppointment')) {
        const urlParams = new URLSearchParams(window.location.search);
        const date: string = urlParams.get('date');
        const conditionId: string = urlParams.get('conditionId');
        element = (
          <xcastven-xkilian-project-appointment-scheduler
            navigate={navigate}
            api={this.api}
            user={user}
            initialDate={date ? new Date(date) : null}
            conditionId={conditionId}
          />
        );
      } else if (this.relativePath.startsWith('account')) {
        element = <xcastven-xkilian-project-account navigate={navigate} />;
      } else {
        element = <xcastven-xkilian-project-not-found navigate={navigate} />;
      }
    }

    return <StyledHost>{element}</StyledHost>;
  }
}

import { Navigate } from '../../utils/types';
import { Component, h, Prop } from '@stencil/core';

type MenuButton = {
  title: string;
  icon: string;
  onClick: () => void;
};

@Component({
  tag: 'xcastven-xkilian-project-menu',
  shadow: false,
})
export class Menu {
  @Prop() navigate: Navigate;
  @Prop() isDoctor: boolean;
  @Prop() isMenuOpen: boolean;
  @Prop() handleResetMenu: () => void;

  private homeButton: MenuButton = {
    title: 'Home',
    icon: 'home',
    onClick: () => {
      this.handleResetMenu();
      this.navigate('./homepage');
    },
  };

  private scheduleAppointmentButton: MenuButton = {
    title: 'Schedule an appointment',
    icon: 'event',
    onClick: () => {
      this.handleResetMenu();
      this.navigate('./scheduleAppointment');
    },
  };

  private registerConditionButton: MenuButton = {
    title: 'Register a condition',
    icon: 'coronavirus',
    onClick: () => {
      this.handleResetMenu();
      this.navigate('./registerCondition');
    },
  };

  private accountButton: MenuButton = {
    title: 'Account',
    icon: 'account_circle',
    onClick: () => {
      this.handleResetMenu();
      this.navigate('./account');
    },
  };

  private logoutButton: MenuButton = {
    title: 'Logout',
    icon: 'logout',
    onClick: () => {
      this.handleResetMenu();
      sessionStorage.setItem('user', null);
      this.navigate('./login');
    },
  };

  private patientMenuButtons: Array<MenuButton> = [
    this.homeButton,
    this.scheduleAppointmentButton,
    this.registerConditionButton,
    this.accountButton,
    this.logoutButton,
  ] satisfies Array<MenuButton>;

  private doctorMenuButtons: Array<MenuButton> = [
    this.homeButton,
    this.accountButton,
    this.logoutButton,
  ] satisfies Array<MenuButton>;

  render() {
    const menuButtons: Array<MenuButton> = this.isDoctor
      ? this.doctorMenuButtons
      : this.patientMenuButtons;

    return (
      <div
        class={`fixed top-0 left-0 z-100 h-full max-w-xs min-w-xs transform bg-white shadow-lg transition-transform duration-300 ${
          this.isMenuOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
        onClick={(e: MouseEvent) => e.stopPropagation()}
      >
        <div class="flex h-full flex-col overflow-y-auto">
          <h2 class="w-full bg-[#7357be] px-4 py-3 text-center text-lg font-medium text-white">
            Menu
          </h2>
          <div class="flex w-full flex-col items-center justify-center p-3">
            {menuButtons.map((menuButton: MenuButton) => (
              <div
                role="button"
                class="flex w-full cursor-pointer flex-row items-center justify-between rounded-full border-2 border-transparent px-4 py-3 hover:border-[#d8c7ed]"
                onClick={menuButton.onClick}
              >
                <span class="font-medium text-gray-600">{menuButton.title}</span>
                <md-icon class="text-gray-600">{menuButton.icon}</md-icon>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }
}

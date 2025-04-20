import { instanceOfDoctor, User } from '../../api/generated';
import { Navigate } from '../../utils/types';
import { formatSpecialization } from '../../utils/utils';
import { StyledHost } from '../StyledHost';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-account',
  shadow: false,
})
export class Account {
  @Prop() navigate: Navigate;
  private user: User = JSON.parse(sessionStorage.getItem('user'));

  private handleBack = () => {
    window.navigation.back();
  };

  private handleLogOut = () => {
    this.navigate('./login');

    sessionStorage.setItem('user', null);
  };

  render() {
    if (!this.user) return;

    return (
      <StyledHost class="flex h-screen w-full flex-col overflow-hidden bg-gray-300">
        <xcastven-xkilian-project-header
          navigate={this.navigate}
          type="account"
          isDoctor={this.user.role === 'doctor'}
        />

        {/* Content */}
        <div class="relative flex h-full w-full flex-1 items-center justify-center">
          <div class="mx-6 w-full max-w-md rounded-md bg-white p-6 shadow-lg md:mx-0">
            <div class="mb-6 flex w-full items-center justify-center">
              {instanceOfDoctor(this.user) ? (
                <img
                  src={'https://cdn-icons-png.freepik.com/256/1021/1021799.png?semt=ais_hybrid'}
                  alt={'Doctor'}
                  class="h-32 w-32 object-cover"
                />
              ) : (
                <img
                  src={'https://cdn-icons-png.flaticon.com/512/1430/1430402.png'}
                  alt={'Patient'}
                  class="h-32 w-32 object-cover"
                />
              )}
            </div>

            <div class="flex flex-row items-center justify-between">
              <span class="text-gray-600">First name:</span>
              <span class="text-base font-medium text-[#7357be]">{this.user.firstName}</span>
            </div>
            <div class="flex flex-row items-center justify-between">
              <span class="text-gray-600">Last name:</span>
              <span class="text-base font-medium text-[#7357be]">{this.user.lastName}</span>
            </div>
            <div class="flex flex-row items-center justify-between">
              <span class="text-gray-600">Email:</span>
              <span class="text-base font-medium text-[#7357be]">{this.user.email}</span>
            </div>
            <div
              class={`flex flex-row items-center justify-between ${!instanceOfDoctor(this.user) && `mb-6`}`}
            >
              <span class="text-gray-600">Role:</span>
              <span class="text-base font-medium text-[#7357be]">
                {this.user.role[0].toUpperCase() + this.user.role.slice(1)}
              </span>
            </div>
            {instanceOfDoctor(this.user) && (
              <div class="mb-6 flex flex-row items-center justify-between">
                <span class="text-gray-600">Specialization:</span>
                <span class="text-base font-medium text-[#7357be]">
                  {formatSpecialization(this.user.specialization)}
                </span>
              </div>
            )}

            <div class="flex flex-row items-center justify-between gap-x-3">
              <md-outlined-button class="w-1/2" onClick={this.handleBack}>
                Back
              </md-outlined-button>
              <md-outlined-button class="w-1/2" onClick={this.handleLogOut}>
                Log out
              </md-outlined-button>
            </div>
          </div>
        </div>
      </StyledHost>
    );
  }
}

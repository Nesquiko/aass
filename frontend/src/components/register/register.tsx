import { Api } from '../../api/api';
import { Registration, SpecializationEnum } from '../../api/generated';
import { Navigate } from '../../utils/types';
import { formatSpecialization } from '../../utils/utils';
import { StyledHost } from '../StyledHost';
import { toastService } from '../services/toast-service';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-register',
  shadow: false,
})
export class Register {
  @Prop() navigate: Navigate;
  @Prop() api: Api;

  @State() email: string = '';
  @State() firstName: string = '';
  @State() lastName: string = '';
  @State() isDoctor: boolean = false;
  @State() specialization: SpecializationEnum = null;

  @State() emailError: string = null;
  @State() firstNameError: string = null;
  @State() lastNameError: string = null;
  @State() specializationError: string = null;

  private handleEmailChange = (event: Event) => {
    this.email = (event.target as HTMLTextAreaElement).value;
  };

  private handleFirstNameChange = (event: Event) => {
    this.firstName = (event.target as HTMLTextAreaElement).value;
  };

  private handleLastNameChange = (event: Event) => {
    this.lastName = (event.target as HTMLTextAreaElement).value;
  };

  private handleSpecializationChange = (event: Event) => {
    this.specialization = (event.target as HTMLSelectElement).value as SpecializationEnum;
  };

  private handleRegister = async () => {
    this.emailError = null;
    this.firstNameError = null;
    this.lastNameError = null;
    this.specializationError = null;

    if (!this.email) {
      this.emailError = 'Email is required';
    }

    if (!this.firstName) {
      this.firstNameError = 'First name is required';
    }

    if (!this.lastName) {
      this.lastNameError = 'Last name is required';
    }

    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (this.email && !emailRegex.test(this.email)) {
      this.emailError = 'Invalid email format';
    }

    if (this.isDoctor && !this.specialization) {
      this.specializationError = 'Specialization is required';
    }

    if (this.emailError || this.firstNameError || this.lastNameError || this.specializationError) {
      return;
    }

    let request: Registration;
    if (this.isDoctor) {
      request = {
        role: 'doctor',
        email: this.email,
        firstName: this.firstName,
        lastName: this.lastName,
        specialization: this.specialization,
      };
    } else {
      request = {
        role: 'patient',
        email: this.email,
        firstName: this.firstName,
        lastName: this.lastName,
      };
    }

    try {
      await this.api.auth.registerUser({ registration: request });
      this.navigate('./login');
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  render() {
    return (
      <StyledHost class="flex h-screen w-full flex-row items-center justify-center overflow-x-hidden overflow-y-auto bg-gray-300">
        <div class="mx-6 w-full max-w-md rounded-md bg-white shadow-lg md:mx-0">
          <h1 class="mb-3 w-full rounded-t-lg bg-[#7357be] px-4 py-3 text-center text-2xl">
            <span class="font-normal text-gray-200">Register at </span>
            <span class="font-medium text-white">MediCal</span>
          </h1>
          <div class="w-full p-6">
            <md-filled-text-field
              label="Email"
              class="mb-6 w-full"
              value={this.email}
              onInput={(e: Event) => this.handleEmailChange(e)}
            />
            <div class="flex flex-row items-center justify-between gap-x-3">
              <md-filled-text-field
                label="First Name"
                class="mb-6 w-full"
                value={this.firstName}
                onInput={(e: Event) => this.handleFirstNameChange(e)}
              />
              <md-filled-text-field
                label="Last Name"
                class="mb-6 w-full"
                value={this.lastName}
                onInput={(e: Event) => this.handleLastNameChange(e)}
              />
            </div>

            {this.emailError ? (
              <div class="mb-6 w-full text-center text-sm text-red-500">{this.emailError}</div>
            ) : this.firstNameError ? (
              <div class="mb-6 w-full text-center text-sm text-red-500">{this.firstNameError}</div>
            ) : (
              this.lastNameError && (
                <div class="mb-6 w-full text-center text-sm text-red-500">{this.lastNameError}</div>
              )
            )}

            <div class="mb-6 flex w-full flex-col items-center justify-center gap-y-2">
              <label class="w-full text-center text-sm font-medium text-gray-400">
                Register as
              </label>
              <div class="relative flex w-full max-w-xs cursor-pointer rounded-full border-2 border-[#d8c7ed] p-0.5">
                <div
                  class={`absolute top-0 h-full w-1/2 rounded-full bg-[#7357be] transition-all duration-300 ease-in-out ${
                    !this.isDoctor ? 'left-0' : 'left-1/2'
                  }`}
                ></div>

                <div
                  class={`relative z-10 flex-1 rounded-full py-1 text-center transition-opacity duration-300 ${
                    !this.isDoctor ? 'text-white' : 'text-[#7357be] hover:bg-gray-100'
                  }`}
                  onClick={() => (this.isDoctor = false)}
                >
                  Patient
                </div>

                <div
                  class={`relative z-10 flex-1 rounded-full py-1 text-center transition-opacity duration-300 ${
                    this.isDoctor ? 'text-white' : 'text-[#7357be] hover:bg-gray-100'
                  }`}
                  onClick={() => (this.isDoctor = true)}
                >
                  Doctor
                </div>
              </div>
            </div>

            <div class="mb-6 w-full animate-[slideInFromRight_0.5s_ease-out]">
              <md-filled-select
                label="Specialization"
                class="w-full"
                value={this.specialization}
                onInput={(e: Event) => this.handleSpecializationChange(e)}
                disabled={!this.isDoctor}
              >
                {Object.values(SpecializationEnum).map((specialization: SpecializationEnum) => (
                  <md-select-option value={specialization}>
                    <div slot="headline">{formatSpecialization(specialization)}</div>
                  </md-select-option>
                ))}
              </md-filled-select>
            </div>

            <md-text-button
              class="mb-6 w-full rounded-full"
              onClick={() => this.navigate('./login')}
            >
              Already have an account?
            </md-text-button>

            <md-filled-button
              class="w-full rounded-full bg-[#9d83c6] px-4 py-3"
              onClick={this.handleRegister}
            >
              Register
            </md-filled-button>
          </div>
        </div>
      </StyledHost>
    );
  }
}

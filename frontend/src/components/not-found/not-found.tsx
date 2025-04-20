import { Navigate } from '../../utils/types';
import { StyledHost } from '../StyledHost';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-not-found',
  shadow: false,
})
export class NotFound {
  @Prop() navigate: Navigate;
  render() {
    return (
      <StyledHost class="flex h-screen w-full flex-col overflow-hidden bg-gray-300">
        <div class="relative flex h-full w-full flex-1 items-center justify-center">
          <div class="mx-6 w-full max-w-md rounded-md bg-white p-6 shadow-lg md:mx-0">
            <div class="flex max-w-sm flex-col items-center justify-center gap-y-3 font-medium text-gray-600">
              <md-icon class="h-full w-full" style={{ fontSize: '80px' }}>
                error
              </md-icon>
              <h1 class="mb-6 text-center text-2xl">
                The page you were looking for was not found.
              </h1>
              <md-text-button
                class="w-full rounded-full text-[#7357be]"
                onClick={() => this.navigate('homepage')}
              >
                Back to homepage
              </md-text-button>
            </div>
          </div>
        </div>
      </StyledHost>
    );
  }
}

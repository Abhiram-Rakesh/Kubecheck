{-# LANGUAGE OverloadedStrings #-}

module Rules
    ( checkNoLatestImage
    , checkRequireResources
    , checkNoRootContainers
    ) where

import Data.Text (Text)
import qualified Data.Text as T
import Types

-- | Rule: Disallow 'latest' image tags
-- Severity: ERROR
-- Rationale: Using 'latest' makes deployments non-deterministic
checkNoLatestImage :: Container -> [Violation]
checkNoLatestImage container =
    if hasLatestTag (containerImage container)
        then [Violation SeverityError msg "no-latest-image"]
        else []
  where
    msg = "Container '" <> containerName container <> "' uses 'latest' image tag"
    
    hasLatestTag :: Text -> Bool
    hasLatestTag image =
        let parts = T.splitOn ":" image
        in case parts of
            [_]         -> True  -- No tag means implicit :latest
            [_, tag]    -> tag == "latest"
            _           -> False

-- | Rule: Require resource requests and limits
-- Severity: WARN
-- Rationale: Resources should be set for proper scheduling and stability
checkRequireResources :: Container -> [Violation]
checkRequireResources container =
    case containerResources container of
        Nothing -> [missingResourcesViolation]
        Just res -> checkResourcesComplete res
  where
    missingResourcesViolation =
        Violation SeverityWarn msg "require-resources"
      where
        msg = "Container '" <> containerName container <> "' missing resource requests/limits"

    checkResourcesComplete :: Resources -> [Violation]
    checkResourcesComplete res =
        let requestsViolations = case requests res of
                Nothing -> [Violation SeverityWarn reqMsg "require-resources"]
                Just req -> checkRequirements req "requests"
            limitsViolations = case limits res of
                Nothing -> [Violation SeverityWarn limMsg "require-resources"]
                Just lim -> checkRequirements lim "limits"
        in requestsViolations ++ limitsViolations
      where
        reqMsg = "Container '" <> containerName container <> "' missing resource requests"
        limMsg = "Container '" <> containerName container <> "' missing resource limits"

    checkRequirements :: ResourceRequirements -> Text -> [Violation]
    checkRequirements reqs reqType =
        let cpuMissing = case cpu reqs of
                Nothing -> [Violation SeverityWarn cpuMsg "require-resources"]
                Just _  -> []
            memMissing = case memory reqs of
                Nothing -> [Violation SeverityWarn memMsg "require-resources"]
                Just _  -> []
        in cpuMissing ++ memMissing
      where
        cpuMsg = "Container '" <> containerName container <> "' missing CPU " <> reqType
        memMsg = "Container '" <> containerName container <> "' missing memory " <> reqType

-- | Rule: Containers should not run as root
-- Severity: ERROR
-- Rationale: Running as root increases attack surface
checkNoRootContainers :: Container -> [Violation]
checkNoRootContainers container =
    case containerSecurityContext container of
        Nothing -> [noSecurityContextViolation]
        Just ctx -> checkSecurityContext ctx
  where
    noSecurityContextViolation =
        Violation SeverityError msg "no-root-containers"
      where
        msg = "Container '" <> containerName container 
              <> "' missing securityContext (should set runAsNonRoot: true)"

    checkSecurityContext :: SecurityContext -> [Violation]
    checkSecurityContext ctx =
        case runAsNonRoot ctx of
            Nothing    -> [Violation SeverityError notSetMsg "no-root-containers"]
            Just False -> [Violation SeverityError falseMsg "no-root-containers"]
            Just True  -> checkRunAsUser ctx
      where
        notSetMsg = "Container '" <> containerName container 
                    <> "' has runAsNonRoot not set (should be true)"
        falseMsg = "Container '" <> containerName container 
                   <> "' has runAsNonRoot set to false"

    checkRunAsUser :: SecurityContext -> [Violation]
    checkRunAsUser ctx =
        case runAsUser ctx of
            Nothing -> []
            Just 0  -> [Violation SeverityError rootUserMsg "no-root-containers"]
            Just _  -> []
      where
        rootUserMsg = "Container '" <> containerName container 
                      <> "' explicitly runs as root user (UID 0)"
